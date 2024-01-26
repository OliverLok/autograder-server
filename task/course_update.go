package task

import (
    "errors"
    "fmt"

    "github.com/eriq-augustine/autograder/db"
    "github.com/eriq-augustine/autograder/docker"
    "github.com/eriq-augustine/autograder/lms/lmssync"
    "github.com/eriq-augustine/autograder/log"
    "github.com/eriq-augustine/autograder/model"
    "github.com/eriq-augustine/autograder/model/tasks"
)

func RunCourseUpdateTask(course *model.Course, rawTask tasks.ScheduledTask) (bool, error) {
    task, ok := rawTask.(*tasks.CourseUpdateTask);
    if (!ok) {
        return false, fmt.Errorf("Task is not a CourseUpdateTask: %t (%v).", rawTask, rawTask);
    }

    if (task.Disable) {
        return true, nil;
    }

    _, err := updateCourse(course);

    // Do not reschedule, all course tasks were already scheduled.
    return false, err;
}

// See procetures.UpdateCourse().
func updateCourse(course *model.Course) (bool, error) {
    var errs error;

    // Stop any existing tasks.
    // Note that we are one of the tasks being told to stop.
    stopCoursesInternal(course.GetID(), true, false);

    // Update the course.
    newCourse, updated, err := db.UpdateCourseFromSource(course);
    if (err != nil) {
        // On failure, still try and work with the old course.
        log.Error("Failed to update course.", err, log.NewAttr("course-id", course.GetID()));
        errs = errors.Join(errs, err);
    } else {
        // On success, use the new course.
        course = newCourse;
    }

    // Sync the course.
    _, err = lmssync.SyncLMS(course, false, true);
    if (err != nil) {
        log.Error("Failed to sync course with LMS.", err, log.NewAttr("course-id", course.GetID()));
        errs = errors.Join(errs, err);
    }

    // Build images.
    _, buildErrs := course.BuildAssignmentImages(false, false, docker.NewBuildOptions());
    for imageName, err := range buildErrs {
        log.Error("Failed to build image.", err, log.NewAttr("course-id", course.GetID()), log.NewAttr("image", imageName));
        errs = errors.Join(errs, err);
    }

    // Schedule tasks.
    for _, courseTask := range course.GetTasks() {
        err = Schedule(course, courseTask);
        if (err != nil) {
            log.Error("Failed to schedule task.", err, log.NewAttr("course-id", course.GetID()), log.NewAttr("task", courseTask.String()));
            errs = errors.Join(errs, err);
        }
    }

    return updated, errs;
}
