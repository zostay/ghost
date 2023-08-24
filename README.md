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

```
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

All commands provide a `--config=<path>` option for specifying the location of the configuration file. If not specified, the file is located at `$HOME/.ghost.yaml`.

## Secret Commands

All the secret commands will take a `--keeper=<name>` option that will select the secret keeper to perform the action upon. If not given, it will use the `master` secret keeper. If there is no `master` secret keeper, bad stuff happens (mostly lots of whining).

### set

```
ghost set --name=github.com \
    --username=some.email@example.com \
    --prompt \
    --location=Work
```

This will check to see if there is a secret named `github.com`. If that secret exists, it will update it with the given values. If it does not exist, a new secret will be created. As `name` is not guaranteed to be unique, it will result in an error if multiple secrets named `github.com` are discovered. In which case, you will need to specify the secret by `--id` instead to perform the update.

If the secret already exists and you want to change locations, you will need to specify either `--move` or `--copy`.

When updating a password from the command-line, it is recommended that you use `--prompt` to request the password to avoid inadvertently placing the secret on in your shell history.

### get

```
ghost get --name=github.com
ghost get --id=1238588388299
```

Retrieves one or more secrets from the database. Be aware that `name` is not gauranteed to be unique so this may return more than one secret. If you want just one, you may use with the `--one` option to return just the first found or `--id` to specify the ID of the secret to return (though, this has limited utility since these IDs are not user-friendly and might change on write, depending on the secret keeper).

You may find the `--output=password` command useful if using this with scripts. Be sure to also include `--show-password` to enable the password being output.

### delete

```
ghost delete --name=github.com
```

This will delete exactly one secret from the secret keeper. The `name` field is not guaranteed to be unique, so if multiple secrets have the same name, this operation will refuse to complete. You will need to delete by `--id` instead in that case.

## Additional Secret Commands

### enforce-policy

```
ghost enforce-policy myPolicyKeeper
```

This will enforce lifetime policies on a given keeper. If you want to ensure that certain passwords in a store are cleared after some time period, you must employ some policy enforcement mechanism. This is the most direct. It will immediately list all secrets and any that have a last modified time that is too old according to policy, will be deleted.

If using this method, you may want to use cron or some other job running tool to run this command periodically.

### sync

```
ghost sync myLastPass myKeepass
```

This will perform a synchronization process that will copy all secrets in teh first secret keeper to the second. If the `--delete` option is specified, then it will also delete any secrets from the second that are not found in the first.

## List Commands

### list keepers

```
ghost list keepers
```

This command will list all the keeper types available to the command.

### list locations

```
ghost list locations
```

This will list all the location names available to the secret keeper.

Like the secret commands, you may use the `--keeper=<name>` option to specify the keeper configuration to use or omit it to use the `master` keeper configuration.

### list secrets

```
ghost list secrets --location=Work
```

This will list all the secrets in the given location.

Like the secret commands, you may use the `--keeper=<name>` option to specify the keeper configuration to use or omit it to use the `master` keeper configuration.

## Service Commands

### service start

```
ghost service start --keeper=myPasswordService --enforce-all-policies
```

This will start the password service running in the foreground. It can be set to run against any keeper specified with the `--keeper=<name>` setting or it will use the `master` keeper configuration.

The `--enforce-all-policies` option will cause the server to locate all policy secret keepers and enforce all lifetime policies periodically. The period is determined by the value defined in `--enforcement-period`, which defaults to every minute. If you only want to enforce some of your policies this way, you can specify the policies using the `--enforce-policy` option instead.

### service stop

```
ghost service stop
```

This will locate the running ghost service and stop it. It will attempt a graceful stop by default. If you want to ask it to stop immediately you may specify the `--quit` option. If you want to force stop, use the `--kill` option.

## Configuration Commands

All the configuration commands (actually all the commands) will validate the configuration file on start to ensure the configuration is in a reasonable state before modifications are attempted. It will also check that the modified configuration file will be valid upon write. If it won't be after making the changes your request (e.g., you use `ghost config delete` to delete a keeper that some other secret keeper refers to), then it won't write the configuration changes to the file.

### config set

```
ghost config set <keeper-type> <keeper-name> [ flags ]
```

You will need to look at the usage message for each of the keeper type sub-commands for details. Each works a little different. However, there are some common features. All of these commands will add or modify a keeper configuration in the configuration file.

Sometimes an option will be provided in a special `--*-secret` variant. This allows you to specify a `__SECRET__` reference in the configuration for that particular setting from the command-line. To use it, you will pass a colon-separated list of the relevant fields: keeper, secret, and field.

For example, to set the password to a secret reference in the a Keepass secret keeper, you could do something like this:

```
ghost config set keepass myPasswords \
    --path=$HOME/keepass.kdbx \
    --master-password-secret=keyring:keepass:password
```

### config delete

```
ghost config delete <keeper-name>
```

Removes a keeper configuration from the configuration.

### config get

```
ghost config get <keeper-name>
```

This will resolve the configuration and output it on the command line (with password fields elided).

### config list

```
ghost confg list
```

This is like get, but performs the operaiton for every keeper configuration in the file.

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

A **secret keeper** is a configured storage for secrets. The keepers are divided into two groups, primary secret keepers and secondary, which are used to provide additional services to other keepers. 

### Primary Keepers

The following primary secret keeper types are provided:

 * `http` - The http secret keeper accesses secrets provided by the ghost gRPC service. The ghost service can be run with the `ghost service start` command and used to wrap any keeper in the given configuration. As of this writing, the http keeper may only be used on a local machine as all communication is performed over a unix socket.
 * `human` - The human secret keeper provides a means of asking the person at the keyboard to enter a secret. A human keeper is configured with a number of questions, each acting as a secret the user is expected to supply upon request.
 * `keepass` - The Keepass secret keeper loads and stores secrets in a local Keepass database file. You will need to provide the Keepass secret keeper the path to the file as well as the master password for encrypting and decrypting the file.
 * `keyring` - The keyring secret keeper loads and stores passwords in the system keyring. This should work on macOS, Windows, Linux, and BSD. On macOS it accesses the system keyring using the `security` command. Similarly, it uses the Windows OS keyring on Windows. On Linux and BSD, it uses dbus to communicate with whatever secret service is installed, usually GNOME Keyring. 
 * `lastpass` - The LastPass secret keeper uses the LastPass API to access secrets. The secrets are downloaded from the online store and then decrypted locally on get and encrypted locally and set to the online store during set.
 * `low` - The low security secret keeper stores secrets in a local YAML file in plaintext. This is obviously only suitable for secrets that are not very secure or on a system you are very confident in.
 * `memory` - The memory secret keeper will hold a secret encrypted in memory for the duration of the process. Used with the ghost command, this is not very useful. However, it can be useful as a memory store within the ghost service or embedded in an application. The encryption used doesn't guarantee much in the way of safety as the key is also stored in memory, so it may even be considered superfluous.

### Secondary Keepers

The secondary secret keepers exist to provide additional services on top of another secret keeper store. Here is a list of secondary keepers that are provided.

 * `cache` - The cache secret keeper is based on the memory secret keeper and wraps some other keeper. Whenever the keeper is used for getting a secret, the secret is saved locally. A `cache` keeper does not permit any write operations except delete, which just deletes a secret from the cache. It does not delete the secret from the wrapped store. This is another keeper that is not much use outside of either the ghost service or embedded application.
 * `router` - The router secret keeper combined other secret keepers into a single logical keeper. It uses location as the means by which to decide which keeper to use when getting and storing secrets. If a location that does not match any of the configured routes is used, then a default keeper is used to store that secret.
 * `seq` - The sequential secret keeper combines multiple secret keepers into a single logical keeper. When getting secrets, each keeper is checked for that secret in turn and the first secret found to match is returned. When setting, only the first secret keeper in the sequence is modified.

Other secret keepers may be added in the future.

# Example Keeper Configuration

## cache

Caches secrets on get. Does not permit setting, copying, or moving of secrets. Deletes will only remove the secret from the cache, not the wrapped keeper.

```yaml
keepers:
  my-cache:
    type: cache
    keeper: my-other-keeper
    touch_on_read: false
```

**Type:** `cache`

**Required Fields:**

 * `keeper` - The name of the keeper to wrap. This keeper must exist in the configuration.

**Optional Fields:**

 * `touch_on_read` - If true, the last modified time of the secret will be updated every time the secret is read. This is useful for keeping a secret alive in the cache for a longer based on use. The default is false.

## http

Accesses secrets by contacting the ghost service over a local unix socket. The unix socket is automatically discovered. No configuration of this keeper is possible at this time.

```yaml
keepers:
  my-http:
    type: http
```

**Type:** `http`

**Required Fields:**

None

## human

Asks the person at the keyboard to supply the secret by putting up a password dialog.

```yaml
keepers:
    my-human:
      type: human
      questions:
        - id: LastPass Password
          ask_for: [ password ]
          presets: 
            username: some.email@example.com
```

**Type:** `human`

**Required Fields:**

It is permitted to have no `questions` defined.

 * `questions` - A list of questions to ask the user. Each question must have an `id` and a list of `ask_for` fields. The `id` is the name of the secret to ask for. The `ask_for` fields are the fields to ask for. The `ask_for` fields may be any of `password`, `username`, `url`, `location`, `name`, `type`, or other names for additional fields, if supported by the secret keeper. The `presets` are optional and allow for preset values to be injected into the secret.

## keepass

Accesses secrets using a local Keepass database file.

```yaml
keepers:
  my-keepass:
    type: keepass
    path: /home/user/keepass.kdbx
    master_password:
      __SECRET__:
        keeper: pinentry
        secret: personal-keepass.kdbx-master-password
        field: password
```

**Type:** `keepass`

**Required Fields:**

 * `path` - The path to the Keepass database file.
 * `master_password` - The master password for the Keepass database file. This may be a `__SECRET__` reference value.

## keyring

Accesses secrets through the system keyring.

```yaml
keepers:
  my-keyring:
    type: keyring
    service_name: ghost
```

**Type:** `keyring`

**Required Fields:**

 * `service_name` - The name of the service to use in the keyring.

## lastpass

Accesses secrets through a LastPass vault.

```yaml
keepers:
  my-lastpass:
    type: lastpass
    username: some.email@example.com
    password:
      __SECRET__:
        keeper: pinentry
        secret: LastPass Password
        field: password
```

**Type:** `lastpass`

**Required Fields:**

 * `username` - The username to use to access the LastPass vault.
 * `password` - The password to use to access the LastPass vault. This may be a `__SECRET__` reference value.

## low

Stores secrets in a local YAML file in plaintext. This is not secure.

```yaml
keepers:
  my-low:
    type: low
    path: /home/user/ghost.yaml
```

**Type:** `low`

**Required Fields:**

 * `path` - The path to the YAML file to use.

## memory

Stores secrets in memory. This is only useful when embedded in another application or via the ghost gRPC service.

```yaml
keepers:
  my-memory:
    type: memory
```

**Type:** `memory`

**Required Fields:**

None

## policy

Applies policies to the secrets stored in another keeper. Currently, this includes:

 * A lifetime policy which can be used to expire secrets after a certain amount of time.
 * An acceptance policy that can be used to either allow or deny access to secrets based on matching rules.

For the lifetime policy to operate, you must either run the `ghost enforce-policy` command or run the ghost service with either the `--enforce-all-policies` or `--enforce-policy` options. Enforcement works by walking all secrets in a keeper and checking the last modified date of each against the applicable rules and defaults. If the secret is too old, it is deleted. If a secret keeper does not implement the `List*` methods that allow for walking, lifetime cannot be enforced.

```yaml
keepers:
  my-policy:
    type: policy
    keeper: my-other-keeper
    lifetime: 48h
    acceptance: allow
    rules:
      - location: Work
        acceptance: deny
      - location: Pers*
        acceptance: deny
      - username: /\w+@example\.com/
        lifetime: 10m
```

**Type:** `policy`

**Required Fields:**

 * `keeper` - The name of the keeper to wrap. This keeper must exist in the configuration.
 * `acceptance` - The default acceptance policy. This may be `allow` or `deny`.
 * `rules` - The list of matches and rules to apply for each rule. Rules are matched in the order given. See below for details.

**Optional Fields:**

 * `lifetime` - The default lifetime for secrets. This may be a duration string or a number of seconds. If not provided, the lifetime is not limited.

**Rules:**

Each rule must define either an acceptance or lifetime policy:

 * `acceptance` - The acceptance policy for this rule. This may be `allow` or `deny` or `inherit`.
 * `lifetime` - The lifetime for this rule. This may be a duration string or a number of seconds. If not provided, the lifetime is not limited.

Each rule must also provide one or more matching filters:

 * `location` - The location to match.
 * `name` - The name to match.
 * `username` - The username to match.
 * `type` - The type to match.
 * `url` - The URL to match.

Each match may be one of the following types of value:

 * A string. This is matched directly.
 * A glob. This is matched using typical glob pattern rules where `*` matches many characteres and `?` matches one. Primarily useful for matching prefixes or suffixes.
 * A regular expression. This uses the [Google Re2](https://github.com/google/re2/wiki/Syntax) syntax. To use a regular expression the value must be a string that starts with `/` and ends with `/`. For example, `/^foo/` matches any string that starts with "foo".

## router

Routes secrets to other keepers based on location. If a secret is stored in a location that matches a route, the secret is stored in the keeper for that route. If no route matches, the secret is stored in the default keeper. The same is true for retrieval.

```yaml
keepers:
  my-router:
    type: router
    default: my-default-keeper
    routes:
      - locations: [ Personal ]
        keeper: my-personal-keeper
      - locations: [ Work ]
        keeper: my-work-keeper
      - locations: [ API-Keys, Robots ]
        keeper: my-api-keeper
```

**Type:** `router`

**Required Fields:**

 * `default` - The name of the keeper to use for secrets that do not match any route. This keeper must exist in the configuration.
 * `routes` - The list of routes to use. See below for details.

**Routes:**

Each route must define a list of locations and a keeper:

 * `locations` - The list of locations to match. If a secret is stored in a location that matches one of these, the secret is stored in the keeper for this route.
 * `keeper` - The name of the keeper to use for secrets that match this route. This keeper must exist in the configuration.

## seq

Stores secrets in a sequence of keepers. When getting a secret, the first keeper in the sequence that has the secret is used. When setting a secret, the first keeper in the sequence is used. When deleting a secret, the first keeper in the sequence that has the secret is used.

```yaml
keepers:
  my-seq:
    type: seq
    keepers:
      - my-first-keeper
      - my-second-keeper
      - my-third-keeper
```

**Type:** `seq`

**Required Fields:**

 * `keepers` - The list of keepers to use in the sequence. Each keeper must exist in the configuration.

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
