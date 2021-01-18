#!/usr/bin/env bash
set -e

# Match test scenarios
RE_TEST='\^TestScenarios\$.*'
# Only match scenarios and not individual features or steps
RE_SCENARIO='\^TestScenarios\$/\^Feature_[^/]+/Scenario_[^/]+\$'

CMD=( go test -coverpkg=$(go list ./... | grep -v /test | paste -sd "," -) -coverprofile=coverage.txt -covermode=atomic $(go list ./...) -json )

if [ "$#" -ne 2 ]; then
  # Run only BDD tests if we specify BDD_TAGS to run
	if [ -z $BDD_TAGS ]; then
		"${CMD[@]}"
	else
		"${CMD[@]}" -run "TestScenarios"
	fi
else
	PREFIX="-test.run="
	RUN=$(echo $1 | sed 's/-test.run=//')
	if [[ ${RUN} =~ ${RE_TEST} ]] && [[ ! ${RUN} =~ ${RE_SCENARIO} ]]; then
		TEST=$(echo $RUN | tr -d '$^')
		echo "{\"Time\":\"$(date --rfc-3339=ns | tr ' ' 'T')\",\"Action\":\"skip\",\"Package\":\"$2\",\"Test\":\"$(printf %q "$TEST")\",\"Elapsed\":0}"
	else
		if [[ ${RERECORD_FAILED_TESTS} == "true" ]]; then
			RECORD=true "${CMD[@]}" -run "$RUN" "$2"
		else
			"${CMD[@]}" -run "$RUN" "$2"
		fi
	fi
fi
