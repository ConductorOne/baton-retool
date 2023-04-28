FROM gcr.io/distroless/static-debian11:nonroot
ENTRYPOINT ["/baton-retool"]
COPY baton-retool /