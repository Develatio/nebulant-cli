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

To reproduce *nix and windows builds, just run `make reproducible_buildall` and make a diff between both binaries, the distributed one and the just created.

To reproduce MacOS builds, the process is similar but we just remove the signature before test the differences. Supossing you downloaded a release binary called `nebulant-0.2.0-beta-20220207-darwin-arm64` and you already executed `make reproducible_buildall` wich bring the `bin/nebulant-0.2.0-beta-20220209-darwin-arm64` binary, the steps to test the reproducible build are as simple as follow:

```
$ codesign --remove-signature nebulant-0.2.0-beta-20220207-darwin-arm64
$ codesign --remove-signature nebulant-0.2.0-beta-20220209-darwin-arm64
$ xxd nebulant-0.2.0-beta-20220207-darwin-arm64 >rm_signed_hex
$ xxd nebulant-0.2.0-beta-20220209-darwin-arm64 >unsigned_hex
$ diff rm_signed_hex unsigned_hex >differences_hex
$ cat differences_hex
93c93
< 000005c0: 0080 3400 0000 0000 0000 1401 0000 0000  ..4.............
---
> 000005c0: 120c 3400 0000 0000 0000 1401 0000 0000  ..4.............
258,263c258,263
< 00001010: 505a 2d79 3545 544c 4451 5244 4536 4430  PZ-y5ETLDQRDE6D0
< 00001020: 6f44 4a61 2f57 2d36 6650 374c 797a 4c76  oDJa/W-6fP7LyzLv
< 00001030: 6d54 7632 2d6f 4631 422f 4a37 6768 7971  mTv2-oF1B/J7ghyq
< 00001040: 7973 5f34 724e 545f 3438 304d 4e67 2f65  ys_4rNT_480MNg/e
< 00001050: 5178 7a41 784f 6b62 7455 6947 4b66 7773  QxzAxOkbtUiGKfws
< 00001060: 4a63 6e22 0a20 ff00 0000 0000 0000 0000  Jcn". ..........
---
> 00001010: 3959 5854 4f45 6772 7336 7336 354f 3937  9YXTOEgrs6s65O97
> 00001020: 422d 7572 2f71 6b47 7157 3850 5277 4167  B-ur/qkGqW8PRwAg
> 00001030: 4e67 5135 5f4d 754f 362f 4a37 6768 7971  NgQ5_MuO6/J7ghyq
> 00001040: 7973 5f34 724e 545f 3438 304d 4e67 2f4c  ys_4rNT_480MNg/L
> 00001050: 3051 4877 3864 4e75 566f 6c64 7377 735a  0QHw8dNuVoldswsZ
> 00001060: 4439 4b22 0a20 ff00 0000 0000 0000 0000  D9K". ..........
563194c563194
< 00897f90: 0300 0000 0000 0000 3230 3232 3032 3037  ........20220207
---
> 00897f90: 0300 0000 0000 0000 3230 3232 3032 3039  ........20220209
```

The big difference here is the path in wich the binary was builded (commonly your user home path). This path is inserted by go compiler. The second big difference is the build date inserted into the code during the compilation.


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