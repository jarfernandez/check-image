#!/usr/bin/env bash
set -euo pipefail

# GitHub Action entrypoint for check-image.
# Maps action inputs (INPUT_* env vars) to docker run arguments
# and generates structured outputs and step summaries.

readonly REGISTRY="ghcr.io/jarfernandez/check-image"
readonly WORKSPACE="/github/workspace"
readonly ALL_CHECKS=("age" "size" "ports" "registry" "root-user" "healthcheck" "secrets" "labels")

# --- Build Docker image reference ---
readonly IMAGE_REF="${REGISTRY}:${INPUT_VERSION}"

echo "::group::check-image configuration"
echo "Image to validate: ${INPUT_IMAGE}"
echo "check-image version: ${INPUT_VERSION}"
echo "Docker image: ${IMAGE_REF}"
echo "::endgroup::"

# --- Pull the image ---
echo "::group::Pulling check-image image"
if ! docker pull "${IMAGE_REF}"; then
  echo "::error::Failed to pull ${IMAGE_REF}. Verify the version exists."
  exit 2
fi
echo "::endgroup::"

# --- Build check-image CLI arguments ---
CMD_ARGS=("all" "${INPUT_IMAGE}" "--output" "json")

if [[ -n "${INPUT_LOG_LEVEL}" && "${INPUT_LOG_LEVEL}" != "info" ]]; then
  CMD_ARGS+=("--log-level" "${INPUT_LOG_LEVEL}")
fi

if [[ -n "${INPUT_CONFIG}" ]]; then
  CMD_ARGS+=("--config" "${INPUT_CONFIG}")
fi

if [[ -n "${INPUT_SKIP}" ]]; then
  CMD_ARGS+=("--skip" "${INPUT_SKIP}")
fi

if [[ "${INPUT_FAIL_FAST}" == "true" ]]; then
  CMD_ARGS+=("--fail-fast")
fi

if [[ -n "${INPUT_MAX_AGE}" ]]; then
  CMD_ARGS+=("--max-age" "${INPUT_MAX_AGE}")
fi

if [[ -n "${INPUT_MAX_SIZE}" ]]; then
  CMD_ARGS+=("--max-size" "${INPUT_MAX_SIZE}")
fi

if [[ -n "${INPUT_MAX_LAYERS}" ]]; then
  CMD_ARGS+=("--max-layers" "${INPUT_MAX_LAYERS}")
fi

if [[ -n "${INPUT_ALLOWED_PORTS}" ]]; then
  CMD_ARGS+=("--allowed-ports" "${INPUT_ALLOWED_PORTS}")
fi

if [[ -n "${INPUT_REGISTRY_POLICY}" ]]; then
  CMD_ARGS+=("--registry-policy" "${INPUT_REGISTRY_POLICY}")
fi

if [[ -n "${INPUT_LABELS_POLICY}" ]]; then
  CMD_ARGS+=("--labels-policy" "${INPUT_LABELS_POLICY}")
fi

if [[ -n "${INPUT_SECRETS_POLICY}" ]]; then
  CMD_ARGS+=("--secrets-policy" "${INPUT_SECRETS_POLICY}")
fi

if [[ "${INPUT_SKIP_ENV_VARS}" == "true" ]]; then
  CMD_ARGS+=("--skip-env-vars")
fi

if [[ "${INPUT_SKIP_FILES}" == "true" ]]; then
  CMD_ARGS+=("--skip-files")
fi

# --- Handle 'checks' input: compute skip list from complement ---
if [[ -n "${INPUT_CHECKS}" ]]; then
  declare -A requested
  IFS=',' read -ra check_array <<< "${INPUT_CHECKS}"
  for check in "${check_array[@]}"; do
    check="$(echo "${check}" | xargs)"
    requested["${check}"]=1
  done

  derived_skip=""
  for check in "${ALL_CHECKS[@]}"; do
    if [[ -z "${requested[${check}]+_}" ]]; then
      if [[ -n "${derived_skip}" ]]; then
        derived_skip="${derived_skip},${check}"
      else
        derived_skip="${check}"
      fi
    fi
  done

  if [[ -n "${derived_skip}" ]]; then
    # Merge with existing --skip flag if present
    skip_merged=false
    for i in "${!CMD_ARGS[@]}"; do
      if [[ "${CMD_ARGS[${i}]}" == "--skip" ]]; then
        CMD_ARGS[$((i + 1))]="${CMD_ARGS[$((i + 1))]},${derived_skip}"
        skip_merged=true
        break
      fi
    done
    if [[ "${skip_merged}" == "false" ]]; then
      CMD_ARGS+=("--skip" "${derived_skip}")
    fi
  fi
fi

# --- Execute ---
echo "::group::Running check-image"
full_cmd=(
  "docker" "run" "--rm"
  "-v" "${GITHUB_WORKSPACE}:${WORKSPACE}:ro"
  "-w" "${WORKSPACE}"
  "${IMAGE_REF}"
  "${CMD_ARGS[@]}"
)
echo "Command: ${full_cmd[*]}"

exit_code=0
stderr_file="$(mktemp)"
json_output="$("${full_cmd[@]}" 2>"${stderr_file}")" || exit_code=$?

# Display stderr (log output) in the workflow log
if [[ -s "${stderr_file}" ]]; then
  cat "${stderr_file}" >&2
fi
rm -f "${stderr_file}"

# Display JSON output
if [[ -n "${json_output}" ]]; then
  echo "${json_output}"
fi
echo "::endgroup::"

# --- Determine result from exit code ---
case ${exit_code} in
  0) result="passed" ;;
  1) result="failed" ;;
  *) result="error" ;;
esac

# --- Set outputs ---
echo "result=${result}" >> "${GITHUB_OUTPUT}"

{
  echo "json<<CHECK_IMAGE_JSON_EOF"
  echo "${json_output}"
  echo "CHECK_IMAGE_JSON_EOF"
} >> "${GITHUB_OUTPUT}"

# --- Generate step summary ---
{
  echo "## Check Image Results"
  echo ""
  echo "**Image:** \`${INPUT_IMAGE}\`"
  echo "**Result:** ${result}"
  echo ""
} >> "${GITHUB_STEP_SUMMARY}"

# Parse JSON for a richer summary (jq is pre-installed on GitHub runners)
if echo "${json_output}" | jq empty 2>/dev/null; then
  total=$(echo "${json_output}" | jq -r '.summary.total // empty')

  if [[ -n "${total}" ]]; then
    passed_count=$(echo "${json_output}" | jq -r '.summary.passed // 0')
    failed_count=$(echo "${json_output}" | jq -r '.summary.failed // 0')
    errored_count=$(echo "${json_output}" | jq -r '.summary.errored // 0')
    skipped_list=$(echo "${json_output}" | jq -r '.summary.skipped // [] | join(", ")')

    {
      echo "| Metric | Value |"
      echo "|--------|-------|"
      echo "| Total checks | ${total} |"
      echo "| Passed | ${passed_count} |"
      echo "| Failed | ${failed_count} |"
      echo "| Errored | ${errored_count} |"
    } >> "${GITHUB_STEP_SUMMARY}"

    if [[ -n "${skipped_list}" ]]; then
      echo "| Skipped | ${skipped_list} |" >> "${GITHUB_STEP_SUMMARY}"
    fi

    echo "" >> "${GITHUB_STEP_SUMMARY}"
  fi

  # Show details of failed checks
  failed_checks=$(echo "${json_output}" | jq -r '.checks[]? | select(.passed == false) | "- **\(.check)**: \(.message)"')
  if [[ -n "${failed_checks}" ]]; then
    {
      echo "### Failed Checks"
      echo ""
      echo "${failed_checks}"
      echo ""
    } >> "${GITHUB_STEP_SUMMARY}"
  fi
fi

{
  echo "<details><summary>Full JSON Output</summary>"
  echo ""
  echo '```json'
  echo "${json_output}"
  echo '```'
  echo ""
  echo "</details>"
} >> "${GITHUB_STEP_SUMMARY}"

# --- Propagate exit code ---
exit "${exit_code}"
