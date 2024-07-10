# Gwu
Gwu (Generic/Go Web Utility, pronounced "guu-uuh") is a simple web utility for GoLang, designed with generics (Go >=1.18). Gwu is utility-like: simple, concise, and helpful. Most importantly, Gwu is **std-compliant** and with **0 dependencies**.

Gwu does not bloat your project or enforce any framework on you. Instead, it provides simple functionality that helps you focus on important parts and reduce boilerplate code.

Gwu is open to growth and contributions. Fork the project, add/modify/remove code, and open a PR. I look forward to your ideas and improvements.

I’m committed to keeping Gwu simple and pragmatic. For more information, see the [corresponding blog post](https://blog.ioutil.app).

# Usage
`gwu.Handle` returns a standard `http.Handler`. Here’s a minimal example of how to use `gwu.Handle`:

```go
// Initialize the controller with an in-memory store
ctrl := &Controller{store: make(map[string]Poem)}

// Create a handler using gwu.Handle
h := gwu.Handle(gwu.PathVal("name"), ctrl.ByName)

// Register the handler with the mux
mux.Handle("GET /poem/{name}", h)

// Handler function in the Controller
func (c *Controller) ByName(_ context.Context, name string, _ gwu.HandleOpts) (Poem, int, error) {
    poem, err := c.store.PoemByName(name)
    if err != nil {
        return poem, http.StatusNotFound, ErrNotFound
    }

    return poem, http.StatusOK, nil
}
```

For more details, see the [examples directory](examples):
* [Simple In-Memory Poem Store with JSON API](examples/poem).