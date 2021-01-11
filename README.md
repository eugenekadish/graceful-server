Basic server that handles resource cleanup on shutdown. Follows the options pattern for configuration and uses adapters to wrap handlers with middlewares. Only standard libraries are used.

# Development

Setup a development container using a sample _Go_ container from VS Code. Testing is done with _http_ files that can be executed with the VS Code requests extension.

  * https://code.visualstudio.com/docs/remote/containers
  * https://marketplace.visualstudio.com/items?itemName=humao.rest-client

# Links

Lightweight server wrappers that can simplify the code even further:

  * https://github.com/gorilla/mux
  * https://github.com/labstack/echo

More resources on patterns and best practices:

  * https://www.youtube.com/watch?v=PAAkCSZUG1c
  * https://www.youtube.com/watch?v=24lFtGHWxAQ
  * https://medium.com/@matryer/the-http-handler-wrapper-technique-in-golang-updated-bc7fbcffa702
