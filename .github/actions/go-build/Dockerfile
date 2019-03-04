FROM golang:1.11

LABEL "com.github.actions.name"="go-build"
LABEL "com.github.actions.description"="Build go apps"
LABEL "com.github.actions.icon"="terminal"
LABEL "com.github.actions.color"="blue"

LABEL "repository"="https://github.com/tillkahlbrock/libre"
LABEL "homepage"="https://github.com/tillkahlbrock/libre"
LABEL "maintainer"="Till Kahlbrock <till.kahlbrock@gmail.com>"

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

ENTRYPOINT ["/bin/make"]
CMD ["build"]