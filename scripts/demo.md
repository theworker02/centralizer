# Terminal demo

Do not commit large recordings. Regenerate locally with
[asciinema](https://asciinema.org/) or `vhs`.

```bash
# from the repository root
go run ./cmd/centralizer detect ./examples/go-python/analytics
go run ./cmd/centralizer explain ./examples/go-python/analytics
go run ./cmd/centralizer connect ./examples/go-python/analytics
go run ./cmd/centralizer health ./examples/go-python/analytics
```

Suggested `vhs` tape (install vhs separately):

```
Output docs/demo.gif
Set FontSize 16
Set Width 1200
Set Height 700
Type "centralizer detect ./examples/go-python/analytics"
Enter
Sleep 2s
Type "centralizer explain ./examples/go-python/analytics"
Enter
Sleep 2s
```
