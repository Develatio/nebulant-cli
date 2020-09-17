Nebulant CLI
============

- Website: https://nebulant.io
- Documentation: [https://nebulant.io/docs.html](https://nebulant.io/docs.html)

![Nebulant](https://raw.githubusercontent.com/develatio/nebulant-cli/master/logo.png)

The Nebulant CLI tool is a single binary that can be used as a helper for the web editor (providing live-run and autocomplete features) or as a standalone executor of Nebulant blueprint files (suitable for CI/CD environments).

The Nebulant project is a simplet yet powerful UI-based tool that allows you to define and execute a chain of actions. Think about it as a "cloud automation toolkit" that you can use to script and automate actions performed on your cloud services providers, without writing code. 

Actions can be anything, from simple operations such as sleep or print, to execution control with conditional evaluation and loops, to API calls (eg. create an AWS EC2 instance) performed on your favourite cloud services provider. 

Nebulant is an imperative way for controling resources, which means that instead of describing the final result you're willing to obtain, you have the power to define exactly how and when each action should be done.

For more information, see the [website](https://nebulant.io) of the Nebulant.

Documentation
-------------
Documentation is available on the [Nebulant website](https://nebulant.io) at [Docs](https://nebulant.io/docs.html) section.

Quick Start
-----------

Using this tool is very simple.

```
Usage: nebulant [-options] <nebulantblueprint.json>

  -d	Enable server mode at localhost:15678 to use within Nebulant Pipeline Builder.
  -p	Console colors control. -p=<true|false>. (default true)
```


You can choose between server mode:

- `$ ./nebulant -d`

which will be useful to develop your blueprints with the [Nebulant Builder](https://builder.nebulant.io), or production mode:
 
- `$ ./nebulant myblueprint.json`

with which you will run your blueprints indicating only the path to the json file.

You can also run blueprints from your account if you know the UUID. This will download your project blueprint and run it.

- `./nebulant "nebulant://45de9da7-a0af-4236-b168-61834f111f82"`

Reproducible Build
------------------
[Reproducible builds](https://reproducible-builds.org/) *are a set of software development practices that create an independently-verifiable path from source to binary code.*

At Develatio we believe in transparency and we emphasize the safety of our products. For this reason we have included the `make reproducible_buildall` command with which the builds of the binaries officially distributed can be reproduced.


AWS Specific configuration
--------------------------

For AWS use files `~/.aws/config` and `~/.aws/credentials`

* `~/.aws/config` file example content:

```
[profile nebulant-cli-tests]
region=us-west-2
output=json
```

* `~/.aws/credentials` file example content:

```
[nebulant-cli-tests]
aws_access_key_id=your-key-id
aws_secret_access_key=your-key-secret
```

* Environment vars are also allowed. These will take precedence over files.

```
$ export AWS_ACCESS_KEY_ID=AKIAI...
$ export AWS_SECRET_ACCESS_KEY=wJalrX...
$ export AWS_DEFAULT_REGION=us-west-2
```

Nebulant CLI Configuration
--------------------------
To configura your credentials at Nebulant:

* `~/.nebulat/credentials` file example content:

```
{
 	"default": {
 		"auth_token": "TOKENHASH"
 	}
}
```

If you want to add more settings, add them to this file like this:

```
{
 	"default": {
 		"auth_token": "ID:SECRET"
 	},
 	"my_second_conf": {
 		"auth_token": "ID2:SECRET"
 	}
}
```

You can select between the different settings by setting `ACTIVE_CONF_PROFILE`


```
$  export ACTIVE_CONF_PROFILE=my_second_conf
```

Alternatively you can configure environment vars:

```
$ export NEBULANT_TOKEN_ID=356230....
$ export NEBULANT_TOKEN_SECRET=QP1Meei5Cx9N....
```

Environment vars will take precedence over config files.

Contributing
------------

If you find an issue, please report it on the
[issue tracker](https://github.com/develatio/nebulant-cli/issues/new/choose).

License
-------

[GNU Affero General Public License v3](https://github.com/develatio/nebulant-cli/blob/master/LICENSE)