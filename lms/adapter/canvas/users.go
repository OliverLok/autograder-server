package canvas

import (
    "fmt"
    "net/url"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/lms"
    "github.com/eriq-augustine/autograder/util"
)

func (this *CanvasAdapter) FetchUsers() ([]*lms.User, error) {
    this.getAPILock();
    defer this.releaseAPILock();

    apiEndpoint := fmt.Sprintf(
        "/api/v1/courses/%s/users?include[]=enrollments&per_page=%d",
        this.CourseID, PAGE_SIZE);
    url := this.BaseURL + apiEndpoint;

    headers := this.standardHeaders();

    users := make([]*lms.User, 0);

    for (url != "") {
        body, responseHeaders, err := util.GetWithHeaders(url, headers);

        if (err != nil) {
            return nil, fmt.Errorf("Failed to fetch users: '%w'.", err);
        }

        var pageUsers []*User;
        err = util.JSONFromString(body, &pageUsers);
        if (err != nil) {
            return nil, fmt.Errorf("Failed to unmarshal users page: '%w'.", err);
        }

        for _, user := range pageUsers {
            if (user == nil) {
                continue;
            }

            users = append(users, user.ToLMSType());
        }

        url = fetchNextCanvasLink(responseHeaders);
    }

    return users, nil;
}

func (this *CanvasAdapter) FetchUser(email string) (*lms.User, error) {
    this.getAPILock();
    defer this.releaseAPILock();

    apiEndpoint := fmt.Sprintf(
        "/api/v1/courses/%s/search_users?include[]=enrollments&search_term=%s",
        this.CourseID, url.QueryEscape(email));
    url := this.BaseURL + apiEndpoint;

    headers := this.standardHeaders();
    body, _, err := util.GetWithHeaders(url, headers);

    if (err != nil) {
        return nil, fmt.Errorf("Failed to fetch user '%s': '%w'.", email, err);
    }

    var pageUsers []User;
    err = util.JSONFromString(body, &pageUsers);
    if (err != nil) {
        return nil, fmt.Errorf("Failed to unmarshal user page: '%w'.", err);
    }

    if (len(pageUsers) != 1) {
        log.Warn().Str("email", email).Int("num-results", len(pageUsers)).Msg("Did not find exactly one matching user in canvas.");
        return nil, nil;
    }

    return pageUsers[0].ToLMSType(), nil;
}
