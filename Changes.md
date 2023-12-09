WIP  TBD

 * Rename `ghost list keepers` to `ghost list plugins`.
 * Added a few additional tests.
 * Added the `--verbose` flag to `ghost sync`.
 * Fix: A bug in how password worked could result in the password display being stuck on the screen permanently. This has been fixed.
 * Fix: (Experimental) Ghost was using a library for password entry that is no longer maintained. This has been switched to a new input mechanism. This may change again.
 * Fix: The `ghost config list` and `ghost config get` commands failed when secret references failed to fully resolve because they depende upon the secret service or something. Now, it may not list  complete configuration, but shouldn't fail or cause a password propmt just to call these sub-commands.
 * Fix: When `ghost service start` runs, check the service status and perform an appropriate recovery routine to correct errors automatically.
 * Fix: The router service correctly validates default and routes keepers now.

v0.3.0  2023-10-23

 * Add and implement the `GetServiceInfo` endpoint to the gRPC keeper service interface.
 * Add the `ghost service status` command to query the status of the gRPC keeper service, report it's PID, and report the keeper served and the policy enforcement configuration.

v0.2.0  2023-09-04

 * Rewrite the `ghost sync` command to use name, username, and location as the index key.
 * Add the `keeper.Sync` object to help with copying between between keepers with `AddSecret`, `AddLocationSecret`, `AddSecretKeeper`, `CopyTo`, and `DeleteAbsent` methods and the `NewSync` constructor. This provides the implementation for the rewritten `ghost sync` command.
 * Add retries with automatic back-off to the LastPass secret keeper using the same back-off strategy that lastpass-cli uses.
 * Fix: Use "UserName" as the field for the `Username` field in Keepass (previously, "Username" was incorrectly used).

v0.1.0  2023-08-31

 * Add a `keeper.CheckServer()` function for checking to see if the keeper service is currently running.

v0.0.6  2023-08-30

 * Fix: The secret printing function had ambiguous field display logic which has been corrected.

v0.0.5  2023-08-30

 * Fix: Empty last modified time in Keepass will not panic anymore.

v0.0.4  2023-08-30

 * Fix: When secrets are printed by `ghost` commands, hide empty fields and do not panic on empty URL.

v0.0.3  2023-08-30

 * Fix: The `ghost set` handling of `--id` and `--name` has been fixed.

v0.0.2  2023-08-29

 * Fix: Implemented handling of duration decoding in keeper configuration.
 * Fix: Correct problems with looking up keeper configuration during startup.
 * Fix: Do not report error creating PID file when no error occurred.
 * Fix: Correct the gRPC client setup/dialing code.
 * Fix: URL stringification in the gRPC server.
 * Fix: Handle `$HOME` and `~` in `keepass` and `low` keeper paths.
 * Fix: All output was previously going to stderr, when some output needed to be sent to stdout instead. Stdout is now used correctly in several cases.
 * Fix: Correct bugs that prevented the caching keeper from working. Added tests.

v0.0.1  2023-08-28

 * Initial release.
 * Provides the `config delete` sub-command.
 * Provides the `config get` sub-command.
 * Provides the `config list` sub-command.
 * Provides the `config set` sub-command.
 * Provides the `delete` sub-command.
 * Provides the `enforce-policy` sub-command.
 * Provides the `get` sub-command.
 * Provides the `list keepers` sub-command.
 * Provides the `list locations` sub-command.
 * Provides the `list secrets` sub-command.
 * Provides the `random-password` sub-command.
 * Provides the `service start` sub-command.
 * Provides the `service stop` sub-command.
 * Provides the `set` sub-command.
 * Provides the `sync` sub-command.
 * Defines modules for loading and saving configuration.
 * Defines modules for defining plugins (which have to be compiled directly at this time).
 * Defines the secret keeper and secret interfaces and provides a generic implementation of the secret interface.
 * Defines the `cache` secret keeper for caching the contents of other keepers on retrieval.
 * Defines the `http` secret keeper for working with secrets in the custom gRPC service defined by ghost.
 * Includes the implementation of the custom gRPC service.
 * Defines the `human` secret keeper for retrieving secrets from the user at the keyboard.
 * Defines the `keepass` secret keeper for working with secrets stored in a Keepass database.
 * Defines the `keyring` secret keeper for working with secrets stored in the system keyring.
 * Defines the `lastpass` secret keeper for working with secrets stored in a Lastpass vault.
 * Defines the `low` secret keeper for working with secrets stored in a low-security file.
 * Defines the `memory` secret keeper as a reference implementation that stores secrets in memory.
 * Defines the `policy` secret keeper for enforcing lifetime and access policies to other keepers.
 * Defines the `router` secret keeper for routing secret keeper access requests by location.
 * Defines the `seq` secret keeper for stacking secret keepers like layers of an onion.
