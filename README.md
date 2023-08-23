# ghost

This is a secret toolkit written in Go. I use it to backup my online password vault locally and to provide tooling to allow my scripts and such to retrieve secrets without having to store such things in environment files or other ways that make me nervous.

I only use this on local machines that only I have access. I cannot vouch for the safety of transport or security of storage of any aspect of this system. As per the terms of the license, you use this software entirely at your own risk. I make no guarantees or warranties regarding the security or safety of this system.

Suggested use-cases for this software include to:

* Provide command-line access to your Keepass or Lastpass secrets
* Provide command-line access to API keys in a plaintext store
* Provide scripts and other tools access to your passwords.
* Embed the secret handing that Ghost provides inside of Golang apps
* Synchronize data between password stores for backups or other purposes.

# Installation

At this time, no packaged binaries are provided. Installation is from source. You will need to have a Golang compiler to run:

```shell
go install github.com/zostay/ghost@latest
```

# Getting Started

Before you work with any secrets in the system, you will need to install and configure the `ghost` application. The configuration is performed in a file named `.ghost.yaml` in the root of your home directory by default. You may specify a different location using the `--config` option. The configuration is in YAML format and a simple configuration could look something like this:

```yaml
master: mypasswords
keepers:
  mypasswords:
    type: lastpass
    username: some.email@example.com
    password: hunter2 # DO NOT DO THIS!
```

This configures `mypasswords` as the default (the "master") secret keeper. Without this, you'd need to set `--keeper` on every run of `ghost` to select your keeper. It then configures that keeper to use the lastpass secret keeper driver and gives it a username and password. **Storing your master password in plaintext in this configuration is stupid, so don't.** A much better configuration is shown next. However, for those who just want to skip to the end and see the useful bits. With this configuration (assuming the username and password match a real account), you could run these commands:

```shell
% ghost set --name=github.com --username=some.email@example.com --prompt
password: 
% ghost get --name=github.com
github.com:                                                                                     
  ID: 348737777372782
  Location: Personal
  Username: some.email@example.com
  Password: <hidden>                                                                            
  URL: https://github.com/login
  Modified: 2022-10-14 07:23:33 -0500 CDT
  Type:
```

However, rather than using the above configuration, a better one would be something like this:

```yaml
master: mypasswords
keepers:
  mypasswords:
    type: lastpass
    username: some.email@example.com
    password:
      __SECRET__:
        keeper: pinentry
        secret: LastPass Password
        field: password
  pinentry:
    type: human
    questions:
      - id: LastPass Password
        ask_for: [ password ]
```

This configuration is much more secure. Now, when you access your secrets using `ghost`, it will ask you for your "LastPass Password" using a password dialog. Many keys in the configuration can be replaced with a `__SECRET__` value like this to cause that value to be replaced by a lookup from another secret keeper. In this case, the `pinentry` keeper, which uses the `human` secret keeper driver, which just asks the human at the keyboard for the value. For this configuration to work, the `secret` under `__SECRET__` must match to the `id`, the field requested must be set to `password` and the `human` driver needs to be told to ask for the `password` so it has that value to provide.

Entering your master password every time could be a bit annoying if you access passwords often and you have a good long master password. To avoid this, `ghost` can access passwords in your system keyring, which is typically unlocked by you during login, if you are willing to trust your master password to the system keyring. Here is an example of such a configuration:

```yaml
master: myPasswords
keepers:
  systemKeyring:
    type: keyring
    service_name: ghost

  myPasswords:
    type: keepass
    path: /home/user/keepass.kdbx
    master_password:
      __SECRET__:
        keeper: systemKeyring
        secret: personal-keepass.kdbx-master-password
        field: password
```

For the above to work, you will need to run:

```bash
ghost set --name=personal-keepass.kdbx-master-password --prompt
```

Then, enter your master password.

If you do not trust the system keyring, another option is to configure a password service to run to securely cache the master password in memory. Here's an example configuration that does that:

```yaml
master: myPasswords
keepers:
  pinentry:
    type: human
    questions:
      - id: LastPass Password
        ask_for: [ password ]

  tempPolicy:
    type: policy
    keeper: tempStore
    lifetime: 1h

  tempStore:
    keeper: pinentry
    type: cache

  passwordService:
    type: http

  myPasswords:
    type: lastpass
    username: some.email@example.com
    password:
      __SECRET__:
        keeper: passwordService
        secret: LastPass Password
        field: password
```

For the above to work, you will want to arrange to start the password service upon startup or login. Ensure this command runs on startup:

```shell
ghost service start --keeper tempPolicy --enforce-all-policies
```

Then, when you run a command that accesses the `myPasswords` secret keeper, you will be prompted to enter your "LastPass Password". This will be cached in memory for one hour. Communication with the password service is done over a unix socket, so it shouldn't be accessible by other users on the system or over the network. If you want to restart the hour every time you access a secret, you can add the line `touch_on_read: true` to the `tempStore` keeper configuration."

See the `examples` folder of this project for additional sample configurations.

# Command Line

Here is a summary of the command-line commands provided by `ghost`.

## Storing Secrets

You can store a secret using the `ghost set <keeper>` command where `<keeper>` is the name of a configured secret keeper. If omitted, the master secret keeper will be used.

When setting a secret, you may specify which secret to create or update using either the `--id` or the `--name` options. If `--name` is used, be aware that names are not unique, so if one or more secrets with the given name is found, the first found will be updated.

If `--location` is also provided with `--name`, that may be used to help identify which secret to update out of many.

## Retrieving Secrets

You can retrieve secrets using the `ghost get <keeper>` command where `<keeper>` is the name of a configured secret keeper. If omitted, the master secret keeper will be used.

You can retrieve secrets by `--id`, in which case only one secret will be returned (or none if the identified secret cannot be found).

You may also retrieve secrets by `--name`, which may return multiple secrets.

The `-o` or `--output` option may be used to select an output format. The default output is "table". The `--fields` option may be used to select which fields are displayed. Fields other than `id`, `username`, `password`, `url`, `type`, and `modified-time` may be specified by prefixing the fields with `field-`.

## Synchronizing Secrets

You can copy secrets from one secret store to another using the `ghost sync <from-keeper> <to-keeper>` command. All secrets found in `<from-keeper>` will be copied into `<to-keeper>`.

## Ghost Service

The ghost service allows you to run a secret service process as a grpc server on a unix socket. This can allow you to create a secret service with especially sensitive secrets that are stored locally in memory. For example, you could have a master password for something that you want to enter by hand at login, but don't want to enter it every time you need it.

## Other Operations

Additional commands are provided as well:

* `ghost list keepers` will list all the configurable keeper types
* `ghost list locations <keeper>` will list all locations for a keeper.
* `ghost list secrets <keeper> <location>` will list all secrets for a location in a keeper.

# Concepts

The way this is structured is heavily influenced by LastPass and Keepass. 

## Secrets

Basically, each item of data stored is called a **secret**. A secret has a number of fields. The only required field is ID. The only secure field is Password. The list of fields includes:

 * **ID**. Is the unique ID given to the secret in the store. Every secret stored must have a unique ID, but the format or length is not specified by this. That's up to the secret keeper.
 * **Name**. This is a title for the secret stored, describing it's purpose. For example, account information related to Facebook might have a name of "Facebook.com" while your wifi password might have the name "Home WiFi". The name is for human reference and is not secure. Names do not have to be unique and generally should not be expected to be unique.
 * **Username**. This is the username for the account. This is not secure.
 * **Password**. This is the password or secret information stored. This is not secure.
 * **URL**. This is the URL of the account, if any.
 * **Location**. This is the group or folder that the secret belongs to. This may be empty.
 * **Last Modified**. This is the date of last modification. This is usually set to something and is useful when synchronizing secrets between secret keepers.
 * **Type**. This gives information about the type of secret this represents. This may be highly specific to the secret keeper.
 * **Fields**. All other fields are gathered under this heading. There can zero or more fields here. None are secure.

## Keepers

A **secret keeper** is a configured storage for secrets. This may be:

 * An in-memory database (only useful with the ghost service)
 * An grpc server (for accessing secrets provided by the ghost service)
 * The system keychain
 * A Keepass database 
 * A local YAML file (for insecure storage)
 * A LastPass account
 * Intermediate Seq Service
 * Intermediate Router Service

Other service may be added in the future.

### Seq

A **seq** keeper is one that contains a list of other keepers. The keeper within a seq must not contain itself. To explain how this works, let's say we define a seq keeper containing three keepers named A, B, and C in that order.

 * When listing locations for this seq, all locations from all keepers, A, B, and C, in the list will be returned.
 * When listing secrets in a location, secrets in A, B, and C for that location will be returned.
 * When retrieving a secret, the keeper A will be checked first. If A has that secret, it will be returned. If A does not, then B will be checked. If B has the secret, it is returned. Finally, if neither A or B has the requested secret, C is checked and the secret is returned from C if found there.
 * When getting secrets by name, all secret keepers, A, B, and C, with a secret with the given name will return its secrets.
 * When saving, copying, or moving a secret, the write will always happen to the first keeper, in this case A, in the list.
 * When deleting, the deletion will occur from the first matching secret keeper, similar to how retrieving a secret works.

### Router

A **router** keeper is one that maps locations to keepers. It may also have a default keeper. To help understand how this works, let's consider the following router:

```
A: Personal, Work
B: API Keys
C: SSNs, Bank Accounts
D: Everything Else
```

That is, keeper A is used for locations named "Personal" and "Work", B is used for locations named "API Keys", and so on. D is the default keeper.

A router behaves in the following ways:

 * When listing locations for a seq, all locations in the default keeper D are returned as well the mapped locations "Personal", "Work", "API Keys", "SSNs", and "Bank Accounts".
 * When listing secrets in location "Personal", only secrets found in Keeper A will be returned and only those that have location "Personal" in that keeper.
 * When listing secrets in location "Stuff", only secrets found in the default keeper D will be returned and only those that have location "Stuff" in that keeper.
 * When retrieving a secret, the secret can only be returned from the keeper to which it belongs and only if the identified secret belongs to the correct location. For example, if a secret is retrieved and found in keeper C but in location "Stuff", no secret will be returned. If it is found in "SSNs" or "Bank Accounts", it will be returned. If a secret is found in the default keeper D in location "API Keys", it will not be returned, but will be returned if found in location "Stuff".
 * When getting secrets by name, every keeper in the router list will be checked for secrets by that name. However, secrets found in keeper A must be in locations named "Personal" or "Work". Secrets found in B must be in the location named "API Keys". The secrets found in C must be in the locations named "SSNs" or "Bank Accounts". The secrets found in the default keeper D must be found in locations that are not named "Personal", "Work", "API Keys", "SSNs", and "Bank Accounts".
 * When saving, copying, or moving a secret, the destination will be mapped to the keeper for the location. So if a secret is saved with Location "Stuff" it will go to the default keeper D. If a secret is copied to location named "Personal" it will go to keeper A. If a secret is moved to location named "API Keys" from "Stuff", it will be removed from default keeper D and added to keeper B.
 * When deleting a secret, the deletion is only performed if the secret's location matches it's route. If the identified secret is found in keeper A in location named "Personal", it will be deleted. But if instead it's in A at location named "Stuff", it will not be deleted. If found in default keeper D, the opposite is true. It would be deleted from "Stuff"  but not "Personal".

# Developer Tools

Developers might instead prefer to use the Golang code directly. This aims at providing a number of useful tools to that end. You'll want to peruse the godoc for the [github.com/zostay/ghost](https://pkg.go.dev/github.com/zostay/ghost) package for details.

# Copyright and License

Copyright 2023 Andrew Sterling Hanenkamp.

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the “Software”), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
the Software, and to permit persons to whom the Software is furnished to do so,
subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
