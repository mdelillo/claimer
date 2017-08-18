package commands

//failures
const pool_does_not_exist_claim = "%s does not exist"
const pool_not_specified_claim = "no pool specified"
const pool_already_claimed_claim = "%s is already claimed"
const pool_does_not_exist_destroy = "%s does not exist"
const pool_not_specified_destroy = "no pool specified"
const pool_does_not_exist_owner = "%s does not exist"
const pool_is_not_claimed_owner = "%s is not claimed"
const pool_is_not_claimed_release = "%s is not claimed"

const success_owner = "%s was claimed by %s on %s"
const success_status = "*Claimed by you:* %s\n*Claimed by others:* %s\n*Unclaimed:* %s"
const success_claim = "Claimed %s"
const success_create = "Created %s"
const success_help = "Available commands:\n" +
	"```\n" +
	"  claim <env> [<message>]   Claim an unclaimed environment\n" +
	"  create <env>              Create a new environment\n" +
	"  destroy <env>             Destroy an environment\n" +
	"  owner <env>               Show the user who claimed the environment\n" +
	"  release <env>             Release a claimed environment\n" +
	"  status                    Show claimed and unclaimed environments\n" +
	"  help                      Display this message\n" +
	"```"
const success_destroy = "Destroyed %s"
const success_release = "Released %s"
