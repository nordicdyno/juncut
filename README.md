# juncut

json uncut: fixes json which was cut (substringed json) 

## Examples

    $ echo -n '{"j":{"ab": "c", "d' | juncut
    {"j":{"ab":"c","d<FIXED>":"<FIXED>"}}

    $ echo '{"j":{"abc' | juncut
    {"j":{"abc<FIXED>":"<FIXED>"}}

    $ echo '{"j":{"\\' | juncut --pretty
    {
      "j": {
        "\\<FIXED>": "<FIXED>"
      }
    }
    

## Install

    go install github.com/nordicdyno/juncut@latest
