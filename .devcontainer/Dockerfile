FROM mcr.microsoft.com/vscode/devcontainers/go:0-1.20-bullseye

# [Optional] Uncomment this section to install additional OS packages.
RUN export DEBIAN_FRONTEND=noninteractive \
    &&  curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | gpg --dearmor > /usr/share/keyrings/yarn-archive-keyring.gpg \
    && apt-get install apt-transport-https --yes \
    && apt-get update \
    && apt-get -y install --no-install-recommends \
        bash-completion \
        uuid-runtime

USER vscode

RUN go install -v github.com/cweill/gotests/gotests@v1.6.0

# [Optional] Uncomment this line to install global node packages.
# RUN su vscode -c "source /usr/local/share/nvm/nvm.sh && npm install -g <your-package-here>" 2>&1

WORKDIR /workspace
