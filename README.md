Nebulant CLI
============

![Nebulant](https://raw.githubusercontent.com/develatio/nebulant-cli/main/logo.png)

The Nebulant CLI tool is a single binary that can be used as a companion tool
for the web editor (providing live-run and autocomplete features) or as a
standalone executor of Nebulant blueprint files (suitable for CI/CD
environments).

The Nebulant project is a simple yet powerful UI-based tool that allows you to
define and execute a chain of actions. Think of it as "cloud automation
toolkit" that you can use to script and automate actions performed on your cloud
providers, without writing code.

Actions can be anything, from simple operations such as `sleep` or `print`, to
execution control with conditional evaluation and loops, to API calls (e.g.
`create an AWS EC2 instance`) performed on your favorite cloud provider.

Nebulant is an imperative way of controlling resources, which means that instead
of describing the final result you're willing to obtain, you have the power to
define exactly how and when each action should be done.

For more information, see the [website](https://nebulant.app) of the Nebulant.

<br />

📖 Documentation
--------------------------------------------------------------------------------

Find information about how to use the CLI, showcases, supported cloud providers
and much more at
[our docs website](https://nebulant.app/docs/cli/).

<br />

🏁 Quick Start
--------------------------------------------------------------------------------

There are **two** main ways you can use this CLI:

* as a companion tool for the [Nebulant Builder](https://builder.nebulant.app)
* in a CI/CD pipeline

Let's start with the first one:

Running the CLI as a companion tool for the
[Nebulant Builder](https://builder.nebulant.app) will allow you to enjoy
additional features such as:

* real time data retrieval from your cloud providers
* blueprint execution directly from the browser
* faster searching in specific cloud providers pre-fetched datasets (e.g., AWS's
AMIs)

In order to execute this mode, run the CLI tool with the `serve` command:

```shell
$ ./nebulant serve
```

The CLI should start listening at port `15678` in your `localhost`, which will
allow you to connect to it from the
[Nebulant Builder](https://builder.nebulant.app)

TODO: <^ PUT HERE A SHORT VIDEO OF THE ABOVE ^>

The second way that you can use this CLI is by running blueprints directly with
it. You'll most probably want to do this in any of the following situations:

* you've finished creating your blueprint and you want to run it in a CI/CD pipeline
* you found a blueprint in the
[Nebulant Marketplace](https://builder.nebulant.app) and you want to run it.

Executing a blueprint is as simple as:

```shell
$ ./nebulant run -f your_blueprint.nbp
```

You can also fetch and run blueprints directly from the marketplace:

```shell
$ ./nebulant run organization/collection/blueprint
```

Or from your own private organization, which requires you to sign up and create
a token from the [Nebulant Panel](https://builder.nebulant.app) (check the
`Nebulant CLI configuration` section for more info):

```shell
$ export NEBULANT_TOKEN_ID=...
$ export NEBULANT_TOKEN_SECRET=...
$ ./nebulant run organization/collection/blueprint
```

TODO: <^ PUT HERE A SHORT VIDEO OF THE ABOVE ^>

<br />

⚙️ Building locally
--------------------------------------------------------------------------------

If you want to compile the source code yourself, you can follow these steps:

Using Docker:

```shell
$ docker compose -f docker-compose.yml build --no-cache
$ docker compose -f docker-compose.yml run --rm buildenv all
```

This will build the source code for all supported OSs and architectures.

You can build the code for a specific combination of OS and architecture by
replacing `all` with the desired target in the second command. Example:

```shell
$ docker compose -f docker-compose.yml run --rm buildenv linux-amd64
```

Check the table of supported OSs and architectures.

<br />

🖥️ Supported OSs and architectures:
--------------------------------------------------------------------------------

|         | arm | arm64 | 386 | amd64 |
| ------- | --- | ----- | --- | ----- |
| linux   | ✅  |  ✅   | ✅  | ✅   |
| freebsd | ✅  |  ✅   | ✅  | ✅   |
| openbsd | ✅  |  ✅   | ✅  | ✅   |
| windows | ✅  |  ✅   | ✅  | ✅   |
| darwin  | N/A |  ✅   | N/A | ✅   |

<br />

🧰 Reproducible Build
--------------------------------------------------------------------------------

[Reproducible builds](https://reproducible-builds.org/) *are a set of software
development practices that create an independently-verifiable path from source
to binary code.*

At Develatio, we believe in transparency, and we emphasize the safety of our
products. For this reason, we offer you the means to build the source code
yourself and verify that the resulting binaries match the ones that we provide.

To reproduce ***nix** and **windows** builds follow these steps:

1. Clone the repo: `git clone https://github.com/Develatio/nebulant-cli.git`
2. Checkout the version you'd like to build (e.g. `git checkout v0.6.0`)
3. Build the source code (check the `Building locally` section)
4. Run `diff` between the binary that you just built and the binary that we
provide. Make sure that both binaries have the same **version**, **OS** and
**architecture**. You should see no differences, meaning that the binary that
you downloaded contains the exact same code as the binary you just compiled.

To reproduce **darwin** (aka MacOS) builds, the first 3 steps are the same, but
before running the 4th step you need to perform an extra action.
The binaries that we provide are signed with our private certificate, while the
binaries that you can build from the source code aren't, which means that
`diff`ing the darwin binaries will yield differences. You must remove the
signature from the binary that we provide in order to be able to compare both
binaries.

Removing the signature is as easy as:

```shell
$ codesign --remove-signature nebulant-darwin-arm64
$ xxd nebulant-darwin-arm64 > unsigned-nebulant-darwin-arm64
```

Now you should be able to follow the 4th step and see no differences.

<br />

Nebulant CLI Configuration
--------------------------------------------------------------------------------

Once you have created an account and logged in the
[Nebulant Panel](https://builder.nebulant.app) you can create multiple tokens.
Each token gives you full access to all the blueprints your user has access to.
If you're the **administrator** of the organization, that would be **all
blueprints** in the organization. On the contrary, if you're a **member** of an
organization, that would be all **the blueprints** of all the collections you've
been granted access to.

The CLI can store tokens under profiles, which allows you to easily switch
between them. For example, you might have generated multiple users, each one granted
access only to certain collections of blueprints. Or you might have accounts in
multiple organizations.

You can switch to a profile by either setting the following environment
variable:

```shell
$ export NEBULANT_CONF_PROFILE=my_profile
```

Or by interactively selecting the desired profile:

```shell
$ ./nebulant auth profiles <- ?????
```

Alternatively, if you don't want to use profiles, you can set the following
environment variables:

```shell
$ export NEBULANT_TOKEN_ID=...
$ export NEBULANT_TOKEN_SECRET=...
```

Note that environment variables will take precedence over config files.

<br />

🫡 Contributing
--------------------------------------------------------------------------------

If you find an issue, please report it to the
[issue tracker](https://github.com/develatio/nebulant-cli/issues/new).

<br />

📑 License
--------------------------------------------------------------------------------

Copyright (c) Develatio Technologies S.L. All rights reserved.

Licensed under the [MIT](https://github.com/develatio/nebulant-cli/blob/main/LICENSE) license.
