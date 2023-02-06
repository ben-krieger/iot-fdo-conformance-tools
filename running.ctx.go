package main

import (
	"github.com/WebauthnWorks/fdo-fido-conformance-server/tools"
	fdoshared "github.com/WebauthnWorks/fdo-shared"
)

const APIKEY_RESULT_SUBMISSION = "010203040506"
const APIKEY_BUILDS_URL = "https://builds.fidoalliance.org"
const FDO_SERVICE_URL = "http://fdo.tools"
const TOOLS_MODE = fdoshared.CFG_MODE_ONLINE
const FDO_DEV_ENV_DEFAULT = tools.ENV_DEV

const NOTIFY_SERVICE_HOST = "http://localhost:3031"
const NOTIFY_SERVICE_SECRET = "abcdefg"

const GITHUB_OAUTH2_CLIENTID = "abcdefg"
const GITHUB_OAUTH2_CLIENTISECRET = "abcdefg"
const GITHUB_OAUTH2_REDIRECTURL = "http://localhost:3033/api/oauth2/github/callback"

const GOOGLE_OAUTH2_CLIENTID = "abcdefg"
const GOOGLE_OAUTH2_CLIENTISECRET = "abcdefg"
const GOOGLE_OAUTH2_REDIRECTURL = "http://localhost:3033/api/oauth2/google/callback"
