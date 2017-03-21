# Claimer

Claimer is a slack bot for managing locks in a [concourse resource pool](https://github.com/concourse/pool-resource)

<img src="https://cloud.githubusercontent.com/assets/6590106/24130971/28be872e-0dc2-11e7-930a-ff62c9853d0a.png" alt="Claimer demo" width="500" height="658">

## Installation

### Prerequisites

* An API token for the slack bot
  (see [here](https://api.slack.com/bot-users#how_do_i_create_custom_bot_users_for_my_team) for instruction on creating a new bot user)
* The ID of the slack channel that the bot will listen in (you must invite the bot to this channel).
  You can find this by opening the channel in slack and looking at the last portion of the URL.
  For example: `https://<org>.slack.com/messages/<channelId>/`
* A git repo and deploy key for your pool
  (see [here](https://github.com/concourse/pool-resource#git-repository-structure) for repo structure)

  **NOTE:** Claimer can only claim and release pools that contain a single lock
* Golang 1.7+
* `git` and `ssh`

### Compile and run

```bash
mkdir -p $GOPATH/src/github.com/mdelillo
cd $GOPATH/src/github.com/mdelillo
git clone --recursive https://github.com/mdelillo/claimer
cd claimer
go build .
./claimer \
  -apiToken <api-token> \
  -channelId <channel-id> \
  -repoUrl <repo-url> \
  -deployKey <deploy-key>
```

### Deploying to Cloud Foundry

The provided `manifest.yml` and `Procfile` can be used to push Claimer to [Cloud Foundry](https://www.cloudfoundry.org/).

1. Fill in `manifest.yml` with required environment variables
1. Log in to your CF environment
1. Run `cf push`

## Contributing

Be sure all tests pass (`ginkgo -r .`) and code is formatted (`bin/fmt`) before submitting pull requests.

### Prerequisites

The integration tests run by posting in a real slack channel and ensuring that a real git repo is modified.
In order to run them, you'll need to set up or get access to a slack organization and git repo.

#### Slack

1. Create a bot user for claimer (e.g. `@claimer`)
1. Create another bot user for testing (e.g. `@claimer-integration`)
1. Create a channel (e.g. `#test-claimer`) and add both bots to it
1. Create another channel (e.g. `#other-channel`) and add both bots to it

#### Git

1. Create a github repo
1. [Add a deploy key](https://developer.github.com/guides/managing-deploy-keys/#setup-2)
   (make sure it has write access)
1. Create the following directory structure in your repo:
   ```
   .
   ├── pool-1
   │   ├── claimed
   │   └── unclaimed
   │       └── lock-a
   └── pool-2
       ├── claimed
       └── unclaimed
           ├── lock-a
           └── lock-b
   ```
1. Commit all the files and tag the commit with `initial-state`

### Running the tests

1. Export the following environment variables:
   * `CLAIMER_TEST_API_TOKEN`: API Token for `@claimer`
   * `CLAIMER_TEST_BOT_ID`: Bot ID for `@claimer`. You can get this by visiting `https://slack.com/api/auth.test?token=<api-token>`
   * `CLAIMER_TEST_USER_API_TOKEN`: API Token for `@claimer-integration`
   * `CLAIMER_TEST_USERNAME`: Username of your test user (e.g. `claimer-integration`)
   * `CLAIMER_TEST_CHANNEL_ID`: Channel ID for your `#test-claimer` channel
   * `CLAIMER_TEST_OTHER_CHANNEL_ID`: Channel ID for your `#other-channel` channel
   * `CLAIMER_TEST_REPO_URL`: URL of your git repository
   * `CLAIMER_TEST_DEPLOY_KEY`: Deploy token for your git repository

1. Install ginkgo:
   ```bash
   go install github.com/mdelillo/claimer/vendor/github.com/onsi/ginkgo/ginkgo
   ```

1. Run ginkgo:
   ```bash
   $GOPATH/bin/ginkgo -r .
   ```

### Generating fakes

[Counterfeiter](https://github.com/maxbrunsfeld/counterfeiter) fakes are used heavily in the unit tests.
The `bin/generate-fakes` script can be used to regenerate them.

## Known Issues and Limitations

* Only pools with a single lock can be claimed and released
* Claimer does not respond in slack when some errors occur (e.g. when `claim` is called without a pool)
* Claimer can only listen to messages in one channel at a time

## License

Claimer is licensed under the MIT license.