# BGCS-SITE-PROXY

A container that acts as a static site proxy for Google Cloud Storage static sites. This allows for HTTPS and custom domains using [Google Cloud Run](https://cloud.google.com/run?hl=en) as the runtime.

See the [example Terraform](/example/example.tf) as an example for deployment to GCP. The container is available as `squwid/bgcs-site-proxy:v0.1.2` via [Dockerhub](https://hub.docker.com/repository/docker/squwid/bgcs-site-proxy/tags).

See the [blog post](https://blog.squwid.dev/post/bgcs-static-site/).