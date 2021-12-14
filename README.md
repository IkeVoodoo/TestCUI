# TestCUI
Testing CUI in go, using the TCell Library

Dependencis:
- [TCell](https://github.com/gdamore/tcell/) - github.com/gdamore/tcell/v2

Password in the json is hashed with sha256, to check if the password matches hash it, and then compare with the generated password

Code for hashing the password is:

```go
hashed := sha256.Sum256([]byte(pass))
return hex.EncodeToString(hashed[:])
```
