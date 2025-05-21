#!/bin/bash

# Script Defaults
DEBUG=false
SESSION_REVISION=3 # Keep in sync with value in pkg/session/session.go

# Read command line flags
while getopts ":d" o; do
    case "${o}" in
        d)
            DEBUG=true
            ;;
        *)
            ;;
    esac
done
shift $((OPTIND-1))

if [ "$DEBUG" = true ] ; then
    echo "Running script in debug mode..."
fi

read -p 'Nanit Email: ' EMAIL
read -sp 'Nanit Password: ' PASSWORD

# Check if jq is installed
if ! command -v jq &> /dev/null; then
  echo "jq is not installed. Installing..."

  # Install jq using the appropriate package manager for your system
  if [[ -f /etc/os-release ]]; then
    # For systems using systemd
    . /etc/os-release
    case "$ID" in
      debian|ubuntu|raspbian)
        sudo apt-get update && sudo apt-get install -y jq
        ;;
      centos|fedora|rhel)
        sudo yum install -y jq
        ;;
      *)
        echo "Unsupported distribution. Please install jq manually."
        exit 1
        ;;
    esac
  else
    echo "Unable to determine operating system. Please install jq manually."
    exit 1
  fi
fi

# TODO: show json and disable --silent in curl when debug flag is present
LOGIN=$(jq -n --arg email "$EMAIL" --arg password "$PASSWORD" '{email: $email, password: $password, channel: "email"}' | curl --silent --header 'nanit-api-version: 1' --header 'Content-Type: application/json' -d@- https://api.nanit.com/login)

if [ "$DEBUG" = true ] ; then
    echo "LOGIN Result: $LOGIN"
fi

MFA_TOKEN=$(echo $LOGIN | jq .mfa_token)

if [ "$DEBUG" = true ] ; then
    echo "MFA_TOKEN Result: $MFA_TOKEN"
fi

echo -e "\n"

# Instruct users to include quotes because otherwise leading zeros are trimmed.
read -p 'Code (check your email), include quotes (e.g. "0000"): ' MFA_CODE

# TODO: show json and disable --silent in curl when debug flag is present
BODY=$(jq -n --arg email "$EMAIL" --arg password "$PASSWORD" --argjson mfa_code "$MFA_CODE" --argjson mfa_token "$MFA_TOKEN" '{email: $email, password: $password, mfa_token: $mfa_token, mfa_code: $mfa_code, channel: "email"}' | curl --silent --header 'nanit-api-version: 1' --header 'Content-Type: application/json' -d@- https://api.nanit.com/login)

if [ "$DEBUG" = true ] ; then
    echo "BODY Result: $BODY"
fi

REFRESH_TOKEN=$(echo $BODY | jq -r .refresh_token)

if [ "$DEBUG" = true ] ; then
    echo "REFRESH_TOKEN Result: $REFRESH_TOKEN"
fi

SESSION_JSON='{"revision":'$SESSION_REVISION',"authToken":'$MFA_TOKEN',"refreshToken":"'$REFRESH_TOKEN'"}'

if [ "$DEBUG" = true ] ; then
    echo "SESSION_JSON Result: $SESSION_JSON"
fi

echo "$SESSION_JSON" > /data/session.json
