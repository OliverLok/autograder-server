package config

import (
    "slices"
    "strings"

    "github.com/rs/zerolog/log"

    "github.com/eriq-augustine/autograder/util"
)

var seenOptions = make(map[string]*Option);

// Options are a way to access the general configuration in a structured way.
// Options do not themselves hold a value, but just information on how to access the config.
// Options will generally panic if you try to get the incorrect type from them.
// Users should heavily prefer getting config values via Options rather than directly in the config.
type Option struct {
    Key string
    DefaultValue any
    Description string
}

// Create a new option, will panic on failure.
func newOption(key string, defaultValue any, description string) *Option {
    _, ok := seenOptions[key];
    if (ok) {
        log.Fatal().Str("key", key).Msg("Duplicate option key.");
    }

    option := Option{
        Key: key,
        DefaultValue: defaultValue,
        Description: description,
    }

    seenOptions[key] = &option;
    return &option;
}

func (this *Option) Get() any {
    return GetDefault(this.Key, this.DefaultValue);
}

func (this *Option) GetString() string {
    return GetStringDefault(this.Key, this.DefaultValue.(string));
}

func (this *Option) GetInt() int {
    return GetIntDefault(this.Key, this.DefaultValue.(int));
}

func (this *Option) GetBool() bool {
    return GetBoolDefault(this.Key, this.DefaultValue.(bool));
}

func OptionsToJSON() (string, error) {
    options := make([]*Option, 0, len(seenOptions));

    for _, option := range seenOptions {
        options = append(options, option);
    }

    slices.SortFunc(options, func(a *Option, b *Option) int {
        return strings.Compare(a.Key, b.Key);
    });

    return util.ToJSONIndent(options, "", "    ");
}