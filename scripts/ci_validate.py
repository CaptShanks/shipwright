#!/usr/bin/env python3
"""
CI validation for the shipwright repository.
Each subcommand exits with status 1 on failure and prints errors to stderr.
"""
from __future__ import annotations

import argparse
import json
import os
import re
import sys
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
MARKETPLACE_PATH = REPO_ROOT / ".claude-plugin" / "marketplace.json"
MIN_AGENT_BYTES = 1024


def eprint(*args: object) -> None:
    print(*args, file=sys.stderr)


def load_marketplace() -> dict:
    if not MARKETPLACE_PATH.is_file():
        eprint(f"ERROR: Missing marketplace file: {MARKETPLACE_PATH}")
        sys.exit(1)
    try:
        with open(MARKETPLACE_PATH, encoding="utf-8") as f:
            return json.load(f)
    except json.JSONDecodeError as ex:
        eprint(f"ERROR: Invalid JSON in {MARKETPLACE_PATH}: {ex}")
        sys.exit(1)


def cmd_marketplace() -> int:
    data = load_marketplace()
    plugins = data.get("plugins")
    if not isinstance(plugins, list):
        eprint('ERROR: marketplace.json must contain a "plugins" array.')
        return 1
    required = ("name", "source", "description", "version")
    errors = 0
    for i, entry in enumerate(plugins):
        if not isinstance(entry, dict):
            eprint(f"ERROR: plugins[{i}] must be an object, got {type(entry).__name__}.")
            errors += 1
            continue
        for key in required:
            if key not in entry or entry[key] in (None, ""):
                eprint(
                    f'ERROR: plugins[{i}] missing or empty required field "{key}" '
                    f'(plugin name={entry.get("name", "?")!r}).'
                )
                errors += 1
    if errors:
        eprint(f"ERROR: marketplace.json structure validation failed ({errors} issue(s)).")
        return 1
    print("marketplace.json: OK")
    return 0


def cmd_plugins() -> int:
    data = load_marketplace()
    plugins = data.get("plugins", [])
    errors = 0
    for entry in plugins:
        if not isinstance(entry, dict):
            continue
        source = entry.get("source")
        name = entry.get("name", "?")
        if not source:
            continue
        plugin_root = REPO_ROOT / "plugins" / str(source)
        pj = plugin_root / ".claude-plugin" / "plugin.json"
        if not pj.is_file():
            eprint(f'ERROR: [{name}] Missing plugin.json at {pj.relative_to(REPO_ROOT)}')
            errors += 1
            continue
        try:
            with open(pj, encoding="utf-8") as f:
                pjson = json.load(f)
        except json.JSONDecodeError as ex:
            eprint(f"ERROR: [{name}] Invalid JSON in {pj}: {ex}")
            errors += 1
            continue

        for rel in pjson.get("agents") or []:
            if not isinstance(rel, str):
                eprint(f"ERROR: [{name}] agent entry must be a string, got {rel!r}")
                errors += 1
                continue
            target = (plugin_root / rel).resolve()
            try:
                target.relative_to(REPO_ROOT.resolve())
            except ValueError:
                eprint(f"ERROR: [{name}] agent path escapes repo: {rel}")
                errors += 1
                continue
            if not target.is_file():
                eprint(
                    f"ERROR: [{name}] agent file missing: {rel} "
                    f"(expected {target.relative_to(REPO_ROOT)})"
                )
                errors += 1

        for rel in pjson.get("skills") or []:
            if not isinstance(rel, str):
                eprint(f"ERROR: [{name}] skill entry must be a string, got {rel!r}")
                errors += 1
                continue
            target = (plugin_root / rel).resolve()
            try:
                target.relative_to(REPO_ROOT.resolve())
            except ValueError:
                eprint(f"ERROR: [{name}] skill path escapes repo: {rel}")
                errors += 1
                continue
            if not target.is_dir():
                eprint(
                    f"ERROR: [{name}] skill directory missing or not a directory (after "
                    f"resolving symlinks): {rel} -> {target.relative_to(REPO_ROOT)}"
                )
                errors += 1

    if errors:
        eprint(f"ERROR: Plugin structure validation failed ({errors} issue(s)).")
        return 1
    print("Plugin structure: OK")
    return 0


def _parse_frontmatter_name(content: str) -> tuple[str | None, str | None]:
    """Return (error_message, name_value) if frontmatter parses; error if invalid."""
    if not content.strip():
        return ("file is empty", None)
    if not content.lstrip().startswith("---"):
        return ("missing YAML frontmatter (expected leading ---)", None)
    rest = content.lstrip()[3:]
    if rest.startswith("\n"):
        rest = rest[1:]
    elif rest.startswith("\r\n"):
        rest = rest[2:]
    else:
        return ("invalid frontmatter start", None)
    end_match = re.search(r"\n---\s*(?:\n|\r\n)", rest)
    if not end_match:
        return ("unclosed frontmatter (no closing ---)", None)
    fm = rest[: end_match.start()]
    m = re.search(r"^name:\s*(.+)$", fm, re.MULTILINE)
    if not m:
        return ('frontmatter missing required field "name"', None)
    name_val = m.group(1).strip().strip("\"'")
    if not name_val:
        return ('frontmatter "name" is empty', None)
    return (None, name_val)


def cmd_agents() -> int:
    agents_dir = REPO_ROOT / "plugins"
    paths = sorted(agents_dir.glob("*/agents/*.md"))
    errors = 0
    for path in paths:
        rel = path.relative_to(REPO_ROOT)
        try:
            st = path.stat()
        except OSError as ex:
            eprint(f"ERROR: {rel}: cannot stat: {ex}")
            errors += 1
            continue
        if st.st_size == 0:
            eprint(f"ERROR: {rel}: file is empty")
            errors += 1
            continue
        if st.st_size <= MIN_AGENT_BYTES:
            eprint(
                f"ERROR: {rel}: file too small ({st.st_size} bytes); "
                f"expected more than {MIN_AGENT_BYTES} bytes"
            )
            errors += 1
        try:
            text = path.read_text(encoding="utf-8")
        except OSError as ex:
            eprint(f"ERROR: {rel}: cannot read: {ex}")
            errors += 1
            continue
        err, _ = _parse_frontmatter_name(text)
        if err:
            eprint(f"ERROR: {rel}: {err}")
            errors += 1

    if errors:
        eprint(f"ERROR: Agent definition validation failed ({errors} issue(s)).")
        return 1
    print(f"Agent definitions: OK ({len(paths)} file(s))")
    return 0


def _parse_skill_frontmatter(content: str) -> dict[str, str | None]:
    """Parse SKILL.md frontmatter and return {name, description, error}."""
    if not content.strip():
        return {"error": "file is empty", "name": None, "description": None}
    if not content.lstrip().startswith("---"):
        return {"error": "missing YAML frontmatter (expected leading ---)", "name": None, "description": None}
    rest = content.lstrip()[3:]
    if rest.startswith("\n"):
        rest = rest[1:]
    elif rest.startswith("\r\n"):
        rest = rest[2:]
    else:
        return {"error": "invalid frontmatter start", "name": None, "description": None}
    end_match = re.search(r"\n---\s*(?:\n|\r\n)", rest)
    if not end_match:
        return {"error": "unclosed frontmatter (no closing ---)", "name": None, "description": None}
    fm = rest[: end_match.start()]
    name_match = re.search(r"^name:\s*(.+)$", fm, re.MULTILINE)
    desc_match = re.search(r"^description:\s*(.+)$", fm, re.MULTILINE)
    name_val = name_match.group(1).strip().strip("\"'") if name_match else None
    desc_val = desc_match.group(1).strip().strip("\"'") if desc_match else None
    return {"error": None, "name": name_val, "description": desc_val}


def cmd_skills() -> int:
    skills_root = REPO_ROOT / "_skills"
    if not skills_root.is_dir():
        eprint(f"ERROR: Missing _skills directory: {skills_root}")
        return 1
    errors = 0
    count = 0
    for entry in sorted(skills_root.iterdir()):
        if not entry.is_dir():
            continue
        if entry.name.startswith("."):
            continue
        count += 1
        rel = entry.relative_to(REPO_ROOT)
        skill_md = entry / "SKILL.md"
        if not skill_md.is_file():
            eprint(f"ERROR: {rel}: SKILL.md missing")
            errors += 1
            continue
        if skill_md.stat().st_size == 0:
            eprint(f"ERROR: {rel}/SKILL.md is empty")
            errors += 1
            continue
        text = skill_md.read_text(encoding="utf-8", errors="replace")
        parsed = _parse_skill_frontmatter(text)
        if parsed["error"]:
            eprint(f"ERROR: {rel}/SKILL.md: {parsed['error']}")
            errors += 1
        elif not parsed["name"]:
            eprint(f'ERROR: {rel}/SKILL.md: frontmatter missing required field "name"')
            errors += 1
        elif parsed["name"] != entry.name:
            eprint(
                f'ERROR: {rel}/SKILL.md: frontmatter name "{parsed["name"]}" '
                f'does not match directory name "{entry.name}"'
            )
            errors += 1
        if parsed.get("name") and not parsed.get("description"):
            eprint(f'ERROR: {rel}/SKILL.md: frontmatter missing required field "description"')
            errors += 1
        has_subdirs = any(
            d.is_dir() for d in entry.iterdir()
            if not d.name.startswith(".") and d.name != "__pycache__"
        )
        if not has_subdirs:
            eprint(
                f"ERROR: {rel}: must contain at least one subdirectory "
                f"(checklists/, examples/, references/, assets/, scripts/, etc.)"
            )
            errors += 1

    if errors:
        eprint(f"ERROR: Skill definition validation failed ({errors} issue(s)).")
        return 1
    print(f"Skill definitions: OK ({count} skill(s))")
    return 0


def cmd_symlinks() -> int:
    plugins = REPO_ROOT / "plugins"
    errors = 0
    for dirpath, dirnames, filenames in os.walk(plugins, followlinks=False):
        base = Path(dirpath)
        for name in filenames:
            p = base / name
            if p.is_symlink():
                try:
                    resolved = p.resolve()
                except OSError as ex:
                    eprint(f"ERROR: symlink {p.relative_to(REPO_ROOT)}: resolve failed: {ex}")
                    errors += 1
                    continue
                if not resolved.exists():
                    eprint(
                        f"ERROR: broken symlink {p.relative_to(REPO_ROOT)} -> "
                        f"{os.readlink(p)!r} (resolved path does not exist: {resolved})"
                    )
                    errors += 1
        for name in dirnames:
            p = base / name
            if p.is_symlink():
                try:
                    resolved = p.resolve()
                except OSError as ex:
                    eprint(f"ERROR: symlink {p.relative_to(REPO_ROOT)}: resolve failed: {ex}")
                    errors += 1
                    continue
                if not resolved.exists():
                    eprint(
                        f"ERROR: broken symlink {p.relative_to(REPO_ROOT)} -> "
                        f"{os.readlink(p)!r} (resolved path does not exist: {resolved})"
                    )
                    errors += 1

    if errors:
        eprint(f"ERROR: Symlink validation failed ({errors} issue(s)).")
        return 1
    print("Symlinks under plugins/: OK")
    return 0


def cmd_actions() -> int:
    try:
        import yaml  # type: ignore[import-untyped]
    except ImportError:
        eprint("ERROR: PyYAML is required. Install with: pip install pyyaml")
        return 1

    actions_root = REPO_ROOT / "actions"
    if not actions_root.is_dir():
        eprint(f"ERROR: Missing actions directory: {actions_root}")
        return 1
    errors = 0
    for entry in sorted(actions_root.iterdir()):
        if not entry.is_dir():
            continue
        if entry.name.startswith("."):
            continue
        rel = entry.relative_to(REPO_ROOT)
        ay = entry / "action.yml"
        if not ay.is_file():
            eprint(f"ERROR: {rel}: action.yml missing")
            errors += 1
            continue
        try:
            with open(ay, encoding="utf-8") as f:
                yaml.safe_load(f)
        except yaml.YAMLError as ex:  # type: ignore[attr-defined]
            eprint(f"ERROR: {rel}/action.yml: invalid YAML: {ex}")
            errors += 1
        except OSError as ex:
            eprint(f"ERROR: {rel}/action.yml: cannot read: {ex}")
            errors += 1

    if errors:
        eprint(f"ERROR: Action definition validation failed ({errors} issue(s)).")
        return 1
    print("Action definitions: OK")
    return 0


def main() -> None:
    parser = argparse.ArgumentParser(description="Shipwright CI validation")
    sub = parser.add_subparsers(dest="command", required=True)

    sub.add_parser("marketplace", help="Validate .claude-plugin/marketplace.json")
    sub.add_parser("plugins", help="Validate plugin.json and referenced paths")
    sub.add_parser("agents", help="Validate plugins/*/agents/*.md")
    sub.add_parser("skills", help="Validate _skills/* layout")
    sub.add_parser("symlinks", help="Validate symlinks under plugins/")
    sub.add_parser("actions", help="Validate actions/*/action.yml (YAML parse)")

    args = parser.parse_args()
    cmds = {
        "marketplace": cmd_marketplace,
        "plugins": cmd_plugins,
        "agents": cmd_agents,
        "skills": cmd_skills,
        "symlinks": cmd_symlinks,
        "actions": cmd_actions,
    }
    code = cmds[args.command]()
    raise SystemExit(code)


if __name__ == "__main__":
    main()
