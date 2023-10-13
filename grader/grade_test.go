package grader

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
    "testing"

    "github.com/eriq-augustine/autograder/artifact"
    "github.com/eriq-augustine/autograder/config"
    "github.com/eriq-augustine/autograder/docker"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/util"
)

const BASE_TEST_USER = "test_user@test.com";
const TEST_MESSAGE = "";

func TestDockerSubmissions(test *testing.T) {
    if (config.DOCKER_DISABLE.GetBool()) {
        test.Skip("Docker is disabled, skipping test.");
    }

    if (!docker.CanAccessDocker()) {
        test.Fatal("Could not access docker.");
    }

    runSubmissionTests(test, false, true);
}

func TestNoDockerSubmissions(test *testing.T) {
    runSubmissionTests(test, false, false);
}

func runSubmissionTests(test *testing.T, parallel bool, useDocker bool) {
    config.EnableTestingMode(false, true);

    // Directory where all the test courses and other materials are located.
    baseDir := config.COURSES_ROOT.GetString();

    err := LoadCourses()
    if (err != nil) {
        test.Fatalf("Could not load courses: '%v'.", err);
    }

    if (useDocker) {
        _, errs := BuildDockerImages(false, docker.NewBuildOptions());
        if (len(errs) > 0) {
            for imageName, err := range errs {
                test.Errorf("Failed to build image '%s': '%v'.", imageName, err);
            }

            test.Fatalf("Failed to build docker images: '%v'.", err);
        }
    }

    tempDir, err := os.MkdirTemp("", "submission-tests-");
    if (err != nil) {
        test.Fatalf("Could not create temp dir: '%v'.", err);
    }
    defer os.RemoveAll(tempDir);

    testSubmissionPaths, err := util.FindFiles("test-submission.json", baseDir);
    if (err != nil) {
        test.Fatalf("Could not find test results in '%s': '%v'.", baseDir, err);
    }

    if (len(testSubmissionPaths) == 0) {
        test.Fatalf("Could not find any test cases in '%s'.", baseDir);
    }

    gradeOptions := GradeOptions{
        UseFakeSubmissionsDir: true,
        NoDocker: !useDocker,
    };

    failedTests := make([]string, 0);

    for i, testSubmissionPath := range testSubmissionPaths {
        testID := strings.TrimPrefix(testSubmissionPath, baseDir);
        user := fmt.Sprintf("%03d_%s", i, BASE_TEST_USER);

        ok := test.Run(testID, func(test *testing.T) {
            if (parallel) {
                test.Parallel();
            }

            var testSubmission artifact.TestSubmission;
            err := util.JSONFromFile(testSubmissionPath, &testSubmission);
            if (err != nil) {
                test.Fatalf("Failed to load test submission: '%s': '%v'.", testSubmissionPath, err);
            }

            assignment := fetchTestSubmissionAssignment(testSubmissionPath);
            if (assignment == nil) {
                test.Fatalf("Could not find assignment for test submission '%s'.", testSubmissionPath);
            }

            result, _, _, err := Grade(assignment, filepath.Dir(testSubmissionPath), user, TEST_MESSAGE, gradeOptions);
            if (err != nil) {
                test.Fatalf("Failed to grade assignment: '%v'.", err);
            }

            if (!result.Equals(testSubmission.Result, !testSubmission.IgnoreMessages)) {
                test.Fatalf("Actual output:\n---\n%v\n---\ndoes not match expected output:\n---\n%v\n---\n.", result, &testSubmission.Result);
            }

        });

        if (!ok) {
            failedTests = append(failedTests, testID);
        }
    }

    if (len(failedTests) > 0) {
        test.Fatalf("Failed to run submission test(s): '%s'.", failedTests);
    }
}

// Test submission are withing their assignment's directory,
// just check the source dirs for existing courses and assignments.
func fetchTestSubmissionAssignment(testSubmissionPath string) *model.Assignment {
    testSubmissionPath = util.MustAbs(testSubmissionPath);

    for _, course := range GetCourses() {
        if (!util.PathHasParent(testSubmissionPath, filepath.Dir(course.SourcePath))) {
            continue;
        }

        for _, assignment := range course.Assignments {
            if (util.PathHasParent(testSubmissionPath, filepath.Dir(assignment.SourcePath))) {
                return assignment;
            }
        }
    }

    return nil;
}
