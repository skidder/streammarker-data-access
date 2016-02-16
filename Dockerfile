FROM ubuntu:latest
MAINTAINER Scott Kidder <kidder.scott@gmail.com>

# Install CA certificates to enable SSL connections from app
RUN apt-get update \
    && apt-get install -y ca-certificates --no-install-recommends \
    && rm -rf /var/lib/apt/lists/*

# Define mountable directories.
VOLUME ["/data"]

# Define working directory.
WORKDIR /data

# Copy the binary
COPY bin/streammarker-data-access /streammarker-data-access

# Define default command.
CMD ["/streammarker-data-access"]

# Expose ports.
EXPOSE 3000
EXPOSE 3100
