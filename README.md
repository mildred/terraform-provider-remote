Terraform Provider
==================

The remote terraform provider allows provisionning configuration management
resource to remote hosts. This is the equivalent of the `local` terraform
provider, but that works across ssh.

The idea was that traditional configuration management tools were lacking
important features that terraform has. In particular the tfstate is a great
feature that allows to rollback the deployment to its initial state. With
configuration management tools, the only source of truth is your code and when
you modify it, some resources created from previous executions might be left
over, creating some unclean state.

This provider works by connecting via SSH to the remote host, starting bash and
running an internative bash script. The remote host must be equipped with
standard unix tools:

- `bash`
- `base64`
- `cat`
- `chmod`

Future versions of this provider might use other ways to perform actions on the
remote host, such as using SFTP features of the SSH protocol (but this prevents
using `sudo`) or using another interpreter.

Example use of this provider to provision a file on a remote host:

```hcl
provider "remote" {
}

data "remote_ssh_connection" "host" {
  host = "hostname.example.org"
  user = "username"
}

resource "remote_file" "test" {
  conn = data.remote_ssh_connection.host.conn
  filename = "/tmp/test-remote.txt"
  content  = "test from terraform"
}
```

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 0.10.x
-	[Go](https://golang.org/doc/install) 1.11 (to build the provider plugin)

Building The Provider
---------------------

Run `make build` outside GOPATH or with `GO111MODULE=on`

Using the provider
------------------

Place the `terraform-provider-remote` executable in `~/.terraform.d/plugins`

You can do so with `make user-install`

See: https://www.terraform.io/docs/configuration/providers.html#third-party-plugins

Developing the Provider
-----------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.11+ is *required*). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make bin
...
$ $GOPATH/bin/terraform-provider-remote
...
```

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```
