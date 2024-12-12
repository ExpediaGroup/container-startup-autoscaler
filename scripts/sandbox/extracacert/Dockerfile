ARG BASE_IMAGE=kindest/base:v20241108-5c6d2daf

FROM ${BASE_IMAGE}
ARG EXTRA_CA_CERT_FILENAME
COPY ./${EXTRA_CA_CERT_FILENAME} /usr/local/share/ca-certificates/extra-ca-cert.crt
RUN update-ca-certificates
