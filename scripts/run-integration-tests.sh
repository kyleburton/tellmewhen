#!/bin/bash
set -eEuo pipefail

SCRIPT_DIR="$( cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd )"
PROJ_ROOT="$(dirname "$SCRIPT_DIR")"
# NB: depending on your system, $BGPROC_DEFAULT_SLEEP_SECONDS may not be enough
BGPROC_DEFAULT_SLEEP_SECONDS="0.25"

run_test () {
    local tfn desc outcome
    tfn="$1"  ; shift

    desc="$(grep "^# @doc.$tfn" "$0" | cut -f2 -d\|)"

    printf "########################################\n"
    printf "# START: $tname | \"$desc\"\n"
    "$tfn"
    outcome="$?"
    printf "\n# END: %s\n\n" "$desc"

    if [[ $outcome != 0 ]]; then
        echo "FAIL: test failed \$?=$outcome - $tfn | \"$desc\""
        return $outcome
    else
        echo "SUCCESS: $tfn | \"$desc\""
    fi

}

# @doc.test_command_terminates|test that a process exits
test_command_terminates () {
    "$PROJ_ROOT"/tellmewhen process-exits \
                --command="sleep "$BGPROC_DEFAULT_SLEEP_SECONDS"; date" \
                --notify-by-running="printf '\nOK: process exited'"
}

# @doc.test_pid_exits|test that a pid exits
test_pid_exits () {
    (sleep "$BGPROC_DEFAULT_SLEEP_SECONDS" && echo "bg process complete") &
    PID="$!"
    "$PROJ_ROOT"/tellmewhen pid-exits \
                --pid="$PID" \
                --notify-by-running="printf '\nOK: pid-exited: $PID'"
}

# @doc.test_file_exists|test that a file exists (is created)
test_file_exists () {
    FNAME="./file-exists-test.tmpfile"
    test -f "$FNAME" && rm "$FNAME"
    (sleep "$BGPROC_DEFAULT_SLEEP_SECONDS" && touch "$FNAME") &
    "$PROJ_ROOT"/tellmewhen file-exists \
                --file-name="$FNAME" \
                --notify-by-running="echo 'OK: file existed: $FNAME'"
}

# @doc.test_file_removed|test that a file is removed
test_file_removed () {
    FNAME="./file-exists-test.tmpfile"
    test -f "$FNAME" || touch "$FNAME"
    (sleep "$BGPROC_DEFAULT_SLEEP_SECONDS" && rm "$FNAME") &
    "$PROJ_ROOT"/tellmewhen file-removed \
                --file-name="$FNAME" \
                --notify-by-running="echo 'OK: file removed: $FNAME'"
}

# @doc.test_file_updated|test that a file is updated (mtime changes)
test_file_updated () {
    FNAME="./file-exists-test.tmpfile"
    test -f "$FNAME" || touch "$FNAME"
    (sleep "$BGPROC_DEFAULT_SLEEP_SECONDS" && touch "$FNAME") &
    "$PROJ_ROOT"/tellmewhen file-updated \
                --file-name="$FNAME" \
                --notify-by-running="echo 'OK: file updated: $FNAME'"
}


# @doc.test_dir_exists|test that a dir exists (is created)
test_dir_exists () {
    DNAME="./dir-exists-test.tmpdir"
    test -d "$DNAME" && rm -rf "$DNAME"
    (sleep "$BGPROC_DEFAULT_SLEEP_SECONDS" && mkdir "$DNAME") &
    "$PROJ_ROOT"/tellmewhen dir-exists \
                --dir-name="$DNAME" \
                --notify-by-running="echo 'OK: dir existed: $DNAME'"
}

# @doc.test_dir_removed|test that a dir is removed
test_dir_removed () {
    DNAME="./dir-exists-test.tmpdir"
    test -d "$DNAME" || mkdir "$DNAME"
    (sleep "$BGPROC_DEFAULT_SLEEP_SECONDS" && rm -rf "$DNAME") &
    "$PROJ_ROOT"/tellmewhen dir-removed \
                --dir-name="$DNAME" \
                --notify-by-running="echo 'OK: dir removed: $DNAME'"
}

# @doc.test_dir_updated|test that a dir is updated (mtime changes)
test_dir_updated () {
    DNAME="./dir-exists-test.tmpdir"
    test -f "$DNAME" || mkdir "$DNAME"
    (sleep "$BGPROC_DEFAULT_SLEEP_SECONDS" && touch "$DNAME/tmp.file") &
    "$PROJ_ROOT"/tellmewhen dir-updated \
                --dir-name="$DNAME" \
                --notify-by-running="echo 'OK: dir updated: $DNAME'"
}

# TODO: test_process_succeeds
# TODO: test_process_fails

########################################
all_test_fn_names () {
    grep "^test_" "$0" | cut -f1 -d' '
}

main () {
    while [[ "${1:-}" == --* ]]; do
        case "$1" in
            --list)
                for tname in $(all_test_fn_names); do
                    echo "$0 $tname"
                done
                return 0
                ;;
            --help)
                echo "$0 <test-names>"
                return 0
                ;;
            *)
                echo "Error: unrecognized option: '$1'"
                return 1
                ;;
        esac
    done

    if [[ -n "${1:-}" ]]; then
        for tname in $@; do
            run_test "$tname"
        done
        return 0
    fi

    echo "Running all tests"
    for tname in $(all_test_fn_names); do
        run_test "$tname"
    done
}


main "$@"
