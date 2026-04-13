# Naming Conventions Checklist

## Before Committing, Verify

- [ ] Every variable name describes its content, not its type (`userAge` not `intVal`)
- [ ] Every function name starts with a verb describing its action (`fetchOrders` not `orders`)
- [ ] Boolean variables read as assertions (`isReady`, `hasItems`, `canWrite`)
- [ ] Collection variables are plural (`users` not `userList`, `items` not `itemArray`)
- [ ] No single-letter names outside tight loops or idiomatic patterns (`ctx`, `err`, `ok`)
- [ ] No abbreviations that aren't universally understood in the team
- [ ] Constants are named for their meaning, not their value (`maxRetries` not `three`)
- [ ] Types are nouns, interfaces describe capability
- [ ] No Hungarian notation (`strName`, `bIsValid`) -- the type system handles this
- [ ] Acronyms follow consistent casing (`HTTPClient` or `httpClient`, not `HttpClient`)

## Good vs Bad Examples

| Bad | Good | Why |
|-----|------|-----|
| `data` | `userProfile` | Describes what the data actually is |
| `temp` | `unprocessedOrder` | Reveals intent |
| `flag` | `isRetryable` | Reads as a meaningful assertion |
| `list` | `pendingNotifications` | Specific and plural |
| `doStuff()` | `sendWelcomeEmail()` | Describes the action precisely |
| `proc()` | `processPayment()` | Full words, clear intent |
| `handleIt()` | `validateAndStoreOrder()` | But consider splitting this |
| `mgr` | `connectionManager` | No abbreviations |
