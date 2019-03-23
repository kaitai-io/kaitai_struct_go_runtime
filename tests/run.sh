export GOPATH=$GOPATH:$(pwd)

TEST_OUT_DIR="test_out"
ABS_TEST_OUT_DIR="$(pwd)/$TEST_OUT_DIR"
ABS_REPORT_LOG="$ABS_TEST_OUT_DIR/go/report.log"

rm -rf src
mkdir src
cp -r kaitai_struct_tests/spec/go src/spec
cp -r kaitai_struct_tests/src/* src/
cp -r ci_targets/compiled/go/src/test_formats src/test_formats

rm -rf "$TEST_OUT_DIR"
mkdir -p "$TEST_OUT_DIR/go"
rm -f "$TEST_OUT_DIR/go/build.fails"

keep_compiling=1
while [ "$keep_compiling" = 1 ]; do
    if go test -v spec >"$ABS_REPORT_LOG" 2>&1; then
        keep_compiling=0
        cat "$ABS_REPORT_LOG"
    else
        echo "Got error:"
        cat "$ABS_REPORT_LOG"
        if [ -n "$NO_RECOVER" ]; then
            echo "No recovery requested, bailing out"
            exit 1
        fi
        if egrep "^\.\./\.\./compiled/go/.*:[0-9][0-9]*:" "$ABS_REPORT_LOG" >"$ABS_TEST_OUT_DIR/go/err.now"; then
            cat "$ABS_TEST_OUT_DIR/go/err.now" >>"$ABS_TEST_OUT_DIR/go/build.fails"
            sed 's/:.*//' <"$ABS_TEST_OUT_DIR/go/err.now" | sort -u >"$ABS_TEST_OUT_DIR/go/to_delete.now"
            xargs rm <"$ABS_TEST_OUT_DIR/go/to_delete.now"
            echo "Trying to recover..."
            keep_compiling=1
        elif egrep "^src/test_formats/.*:[0-9][0-9]*:" "$ABS_REPORT_LOG" >"$ABS_TEST_OUT_DIR/go/err.now"; then
            cat "$ABS_TEST_OUT_DIR/go/err.now" >>"$ABS_TEST_OUT_DIR/go/build.fails"
            sed 's/:.*//' <"$ABS_TEST_OUT_DIR/go/err.now" | sort -u >"$ABS_TEST_OUT_DIR/go/to_delete.now"
            xargs rm <"$ABS_TEST_OUT_DIR/go/to_delete.now"
            echo "Trying to recover..."
            keep_compiling=1
         elif egrep "^src/spec/.*:[0-9][0-9]*:" "$ABS_REPORT_LOG" >"$ABS_TEST_OUT_DIR/go/err.now"; then
            cat "$ABS_TEST_OUT_DIR/go/err.now" >>"$ABS_TEST_OUT_DIR/go/build.fails"
            sed 's/:.*//' <"$ABS_TEST_OUT_DIR/go/err.now" | sort -u >"$ABS_TEST_OUT_DIR/go/to_delete.now"
            xargs rm <"$ABS_TEST_OUT_DIR/go/to_delete.now"
            echo "Trying to recover..."
            keep_compiling=1
        elif grep -q '^=== RUN' "$ABS_REPORT_LOG"; then
            echo "Tests completed partially..."
            keep_compiling=0
        else
            echo "Unable to recover, bailing out :("
            keep_compiling=0
            exit 1
        fi
    fi
done

cd ../..
go-junit-report <"$ABS_TEST_OUT_DIR/go/report.log" >"$ABS_TEST_OUT_DIR/go/report.xml"

COUNT_TOTAL=$(grep '^=== RUN' "$ABS_REPORT_LOG" | wc -l)
COUNT_FAIL=$(grep '^--- FAIL' "$ABS_REPORT_LOG" | wc -l)
COUNT_PASS=$(grep '^--- PASS' "$ABS_REPORT_LOG" | wc -l)

echo "Totals: $COUNT_TOTAL ran, $COUNT_PASS passed, $COUNT_FAIL failed"
