FROM scratch

COPY /alertmanager-discord /alertmanager-discord

ENTRYPOINT [ "/alertmanager-discord" ]
CMD [ "" ]