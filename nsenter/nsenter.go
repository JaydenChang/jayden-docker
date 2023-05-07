package nsenter

/*
#define _GNU_SOURCE
#include <errno.h>
#include <sched.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
#include <unistd.h>

// if this package is quoted, this function will run automatic
__attribute__((constructor)) void enter_namespace(void)
{
    char *simple_docker_pid;
    // get pid from system environment
    simple_docker_pid = getenv("simple_docker_pid");
    if (simple_docker_pid)
    {
        fprintf(stdout, "got simple docker pid=%s\n", simple_docker_pid);
    }
    else
    {
        fprintf(stdout, "missing simple docker pid env skip nsenter");
        // if no specified pid, the func will exit
        return;
    }

    char *simple_docker_cmd;
    simple_docker_cmd = getenv("simple_docker_cmd");
    if (simple_docker_cmd)
    {
        fprintf(stdout, "got simple docker cmd=%s\n", simple_docker_cmd);
    }
    else
    {
        fprintf(stdout, "missing simple docker cmd env skip nsenter");
        // if no specified cmd, the func will exit
        return;
    }
    int i;
    char nspath[1024];

    char *namespace[] = {"ipc", "uts", "net", "pid", "mnt"};

    for (i = 0; i < 5; i++)
    {
        // create the target path, like /proc/pid/ns/ipc
        sprintf(nspath, "/proc/%s/ns/%s", simple_docker_pid, namespace[i]);
        int fd = open(nspath, O_RDONLY);
		printf("===== %d %s\n", fd, nspath);
        // call sentns and enter the target namespace
        if (setns(fd, 0) == -1)
        {
            fprintf(stderr, "setns on %s namespace failed: %s\n", namespace[i], strerror(errno));
        }
        else
        {
            fprintf(stdout, "setns on %s namespace succeeded\n", namespace[i]);
        }
        close(fd);
    }
    // run command in target namespace
    int res = system(simple_docker_cmd);
    exit(0);
    return;
}
*/

import "C"
