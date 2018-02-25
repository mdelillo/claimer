web: |
  ssh-keyscan -H github.com > ./known_hosts && \
  claimer \
    -apiToken "$API_TOKEN" \
    -channelId "$CHANNEL_ID" \
    -repoUrl "$REPO_URL" \
    -deployKey "$DEPLOY_KEY" \
    -translationFile "$TRANSLATION_FILE"
