# Docker registry retention

This application is meant for removing manifests to be able to delete images from private Docker registry with the assistance of registry's garbage collection. The application reads image tags and desired amount of tags to be left untouched and deletes the manifests of the other tags, so they're collected during the next garbage collection run.

### How to use:
1. Set rules for deletion into the ```config.properties```. For example foo=1, where ```foo``` is the image tag and ```1``` is the amount of images to be left into the registry after purge.
2. Run the application like you would run a regular Go-application.

### Other
- Application was built with Go purely for fun and because I didn't want to script this task. It was written really quickly, so it's not the cleanest code I've written and should be improved, when I have time.e