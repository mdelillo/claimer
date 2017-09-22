package translations

const helpText = "  Available commands:\n" +
	"  ```\n" +
	"    claim <env> [<message>]   Claim an unclaimed environment\n" +
	"    create <env>              Create a new environment\n" +
	"    destroy <env>             Destroy an environment\n" +
	"    notify                    Notify all owners of claimed environments\n" +
	"    owner <env>               Show the user who claimed the environment\n" +
	"    release <env>             Release a claimed environment\n" +
	"    status                    Show claimed and unclaimed environments\n" +
	"    help                      Display this message\n" +
	"  ```"
const DefaultTranslations = `---
claim:
  success: "Claimed {{.pool}}"
  pool_is_already_claimed: "{{.pool}} is already claimed"
  pool_does_not_exist: "{{.pool}} does not exist"
create:
  success: "Created {{.pool}}"
  pool_already_exists: "{{.pool}} already exists"
destroy:
  success: "Destroyed {{.pool}}"
  pool_does_not_exist: "{{.pool}} does not exist"
notify:
  success: "Currently claimed locks, please release if not in use:\n{{.mentions}}"
owner:
  success: "{{.pool}} was claimed by {{.owner}} on {{.date}}"
  pool_does_not_exist: "{{.pool}} does not exist"
  pool_is_not_claimed: "{{.pool}} is not claimed"
release:
  success: "Released {{.pool}}"
  pool_does_not_exist: "{{.pool}} does not exist"
  pool_is_not_claimed: "{{.pool}} is not claimed"
status:
  success: "*Claimed by you:* {{.usersClaimed}}\n*Claimed by others:* {{.otherClaimed}}\n*Unclaimed:* {{.unclaimed}}"
` +
	"unknown_command: \"Unknown command. Try `@claimer help` to see usage.\"\n" +
	"help: |\n" + helpText
