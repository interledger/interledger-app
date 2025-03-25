FROM temporalio/auto-setup:1.18.5

COPY ./temporal-autosetup.sh /etc/temporal/auto-setup.sh

CMD ["autosetup"]

ENTRYPOINT ["/etc/temporal/entrypoint.sh"]
