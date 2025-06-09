FROM debian:bookworm AS debian

ARG VERSION
RUN test -n "${VERSION}" || (echo "Error: VERSION is not set" && exit 1)

ARG REVISION="1"
ARG ARCH
RUN test -n "${ARCH}" || (echo "Error: ARCH is not set" && exit 1)

ARG DEB_NAME
RUN test -n "${DEB_NAME}" || (echo "Error: DEB_NAME is not set" && exit 1)

ARG BINARY_NAME
RUN test -n "${BINARY_NAME}" || (echo "Error: BINARY_NAME is not set" && exit 1)

RUN apt-get update && apt-get install -y sed

ARG GO_BINARY

COPY ./debian/${DEB_NAME} /${DEB_NAME}/
COPY ./${GO_BINARY} /${DEB_NAME}/usr/bin/${BINARY_NAME}
COPY ./debian/${DEB_NAME}/DEBIAN/copyright /${DEB_NAME}/usr/share/doc/copyright

# Go application are tagged with `v` prefix, this remove the first v if present
RUN export VERSION=$(echo "${VERSION}" | sed -e "s/^v\(.*\)/\1/") && \
    sed -i "s/\$ARCH/${ARCH}/" /${DEB_NAME}/DEBIAN/control && \
    sed -i "s/\$VERSION/${VERSION}/" /${DEB_NAME}/DEBIAN/control && \
    dpkg-deb --build --root-owner-group /${DEB_NAME} &&\
    mv /${DEB_NAME}.deb "/${DEB_NAME}_${VERSION}-${REVISION}_${ARCH}.deb"

FROM scratch

COPY --from=debian /*.deb /
