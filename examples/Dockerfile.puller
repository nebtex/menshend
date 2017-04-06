FROM nebtex/menshend:development

# add services

ADD frontend-branch.yml /services/frontend-branch.yml
ADD terminal.yml /services/terminal.yml
ADD frontend-branch-2.yml /services/frontend-branch-2.yml
ADD entrypoint.sh /bin/entrypoint.sh
RUN chmod +x /bin/entrypoint.sh
ENTRYPOINT ["/bin/entrypoint.sh"]
