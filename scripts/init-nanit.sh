#! /bin/bash
SESSION_REVISION=3

read -p 'Nanit Email: ' EMAIL
read -sp 'Nanit Password: ' PASSWORD

LOGIN=$(curl --silent --header 'nanit-api-version: 1' --header 'Content-Type: application/json' --data '{"email":"'$EMAIL'","password":"'$PASSWORD'","channel":"email"}' https://api.nanit.com/login)

MFA_TOKEN=$(echo $LOGIN | jq .mfa_token)

echo -e "\n"

read -p 'Code (check your email): ' MFA_CODE

BODY=$(curl --silent --header 'nanit-api-version: 1' --header 'Content-Type: application/json' --data '{"email":"'$EMAIL'","mfa_code":"'$MFA_CODE'","mfa_token":'$MFA_TOKEN',"password":"'$PASSWORD'"}' https://api.nanit.com/login)

REFRESH_TOKEN=$(echo $BODY | jq -r .refresh_token)

echo '{"revision":'$SESSION_REVISION',"authToken":'$MFA_TOKEN',"refreshToken":"'$REFRESH_TOKEN'"}' > /data/session.json