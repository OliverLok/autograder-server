package grader

import (
	"fmt"
	"testing"

	"github.com/eriq-augustine/autograder/config"
	"github.com/eriq-augustine/autograder/docker"
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
    oldDockerVal := config.DOCKER_DISABLE.GetBool();
    config.DOCKER_DISABLE.Set(true);
    defer config.DOCKER_DISABLE.Set(oldDockerVal);

    runSubmissionTests(test, false, false);
}

func runSubmissionTests(test *testing.T, parallel bool, useDocker bool) {
    config.EnableTestingMode(false, true);

    // Directory where all the test courses and other materials are located.
    baseDir := config.COURSES_ROOT.GetString();

    if (useDocker) {
        _, errs := BuildDockerImages(false, docker.NewBuildOptions());
        if (len(errs) > 0) {
            for imageName, err := range errs {
                test.Errorf("Failed to build image '%s': '%v'.", imageName, err);
            }

            test.Fatalf("Failed to build docker images.");
        }
    }

    gradeOptions := GradeOptions{
        UseFakeSubmissionsDir: true,
        NoDocker: !useDocker,
    };

    testSubmissions, err := GetTestSubmissions(baseDir);
    if (err != nil) {
        test.Fatalf("Error getting test submissions in '%s': '%v'.", baseDir, err);
    }

    if (len(testSubmissions) == 0) {
        test.Fatalf("Could not find any test submissions in '%s'.", baseDir);
    }

    failedTests := make([]string, 0);

    for i, testSubmission := range testSubmissions {
        user := fmt.Sprintf("%03d_%s", i, BASE_TEST_USER);

        ok := test.Run(testSubmission.ID, func(test *testing.T) {
            if (parallel) {
                test.Parallel();
            }

            result, _, _, err := Grade(testSubmission.Assignment, testSubmission.Dir, user, TEST_MESSAGE, gradeOptions);
            if (err != nil) {
                test.Fatalf("Failed to grade assignment: '%v'.", err);
            }

            if (!result.Equals(testSubmission.TestSubmission.Result, !testSubmission.TestSubmission.IgnoreMessages)) {
                test.Fatalf("Actual output:\n---\n%v\n---\ndoes not match expected output:\n---\n%v\n---\n.",
                        result, &testSubmission.TestSubmission.Result);
            }

        });

        if (!ok) {
            failedTests = append(failedTests, testSubmission.ID);
        }
    }

    if (len(failedTests) > 0) {
        test.Fatalf("Failed to run submission test(s): '%s'.", failedTests);
    }
}
