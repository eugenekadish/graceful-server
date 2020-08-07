# Graceful Server

Basic server with handeling of resource clean up when exeting. Following the options patter for configuration and the adapter pattern for middlewares only standard go libraries are used.

  * https://www.youtube.com/watch?v=24lFtGHWxAQ
  * https://medium.com/@matryer/the-http-handler-wrapper-technique-in-golang-updated-bc7fbcffa702

# Development

Full dockerization was done using the example golang image and requests defined with *http* using the Rest Client VS Code Extension:

  * https://code.visualstudio.com/docs/remote/containers
  * https://marketplace.visualstudio.com/items?itemName=humao.rest-client

With VS Code the APIs can be called using the requests extension, but I haven't been able to get it working within the Docker container when working in the remote docker container.

# Links

Some lightweight wrappers that can simplify servers like this even further:

  * https://github.com/gorilla/mux
  * https://github.com/labstack/echo
