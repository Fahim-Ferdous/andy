#!/bin/bash

echo "[*] Test CI Passes"
act --eventpath=gh_actions_pull_request_should_pass.json --rebuild=false --pull=false pull_request

echo "[*] Test CI Fails"
act --eventpath=gh_actions_pull_request_should_fail.json --rebuild=false --pull=false pull_request
