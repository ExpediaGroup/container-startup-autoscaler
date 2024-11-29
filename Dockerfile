# Copyright 2024 Expedia Group, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM --platform=$BUILDPLATFORM golang:1.22 AS build
RUN echo "Build platform: $BUILDPLATFORM, target platform: $TARGETPLATFORM, target OS: $TARGETOS, target arch: $TARGETARCH"
RUN openssl s_client -showcerts -connect proxy.golang.org:443 </dev/null 2>/dev/null|openssl x509 -outform PEM > /usr/local/share/ca-certificates/goproxy.crt
RUN update-ca-certificates
WORKDIR /csa
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-w -s" -o ./csa ./cmd/container-startup-autoscaler

FROM scratch
COPY --from=build /csa/csa /csa/csa
EXPOSE 8080/tcp
EXPOSE 8081/tcp
EXPOSE 8082/tcp
ENTRYPOINT ["/csa/csa"]
