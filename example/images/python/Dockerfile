FROM base
USER 0

ENV PYENV_ROOT=/usr/share/pyenv \
	PATH=/root/.local/bin:/usr/share/pyenv/shims:/usr/share/pyenv/bin:$PATH \
	PYTHON_CONFIGURE_OPTS="--enable-shared --enable-optimizations"

RUN apt-get update \
    && apt-get install -y \
		build-essential \
		libbz2-dev \
		liblzma-dev \
		libncurses5-dev \
		libncursesw5-dev \
		libreadline-dev \
		libffi-dev \
		libsqlite3-dev \
		libssl-dev \
		libxml2-dev \
		libxmlsec1-dev \
		llvm \
		make \
		python3-openssl \
		tk-dev \
		xz-utils \
		zlib1g-dev && \
	curl -sSL "https://github.com/pyenv/pyenv-installer/raw/master/bin/pyenv-installer" | bash \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

ARG PYTHON_VERSION
ENV PYTHON_VERSION="${PYTHON_VERSION}" \
	PIPENV_DEFAULT_PYTHON_VERSION=${PYTHON_VERSION}
RUN pyenv install "${PYTHON_VERSION}" \
    && pyenv global "${PYTHON_VERSION}" \
    && pip install pipx \
    && pipx install pipenv

ARG UV_VERSION
RUN pipx install uv=={{ "uv" | resolve_version}}

ARG POETRY_VERSION
RUN pipx install poetry=={{ "poetry" | resolve_version }}

USER 1000
