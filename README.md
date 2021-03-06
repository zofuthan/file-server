[![Build Status](https://travis-ci.org/cloudfoundry-incubator/file-server.svg)](https://travis-ci.org/cloudfoundry-incubator/file-server)
[![Coverage Status](https://coveralls.io/repos/cloudfoundry-incubator/file-server/badge.png)](https://coveralls.io/r/cloudfoundry-incubator/file-server)
File Server
===========

Bob Loblaw's Blob Store

## Uploading Droplets & Build Artifacts

Uploading droplets & build artifacts via CC involves crafting a correctly-formed multipart request. For Droplets we also poll until the async job completes.

To specify a remote cloud controller to test against, use the following environment variables:

CC_ADDRESS the hostname for a deployed CC
CC_USERNAME, CC_PASSWORD the basic auth credentials for the droplet upload endpoint
CC_APPGUID a valid app guid on that deployed CC

####Learn more about Diego and its components at [diego-design-notes](https://github.com/cloudfoundry-incubator/diego-design-notes)
