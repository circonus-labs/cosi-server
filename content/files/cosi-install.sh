#!/usr/bin/env bash

#
# Optional configuration file
#
cosi_config_file="/etc/default/cosi"

set -o errtrace
set -o errexit
set -o nounset

#
# internal functions
#
usage() {
  printf "%b" "Circonus One Step Install Help

Usage

  ${GREEN}cosi-install --key <apikey> --app <apiapp> [options]${NORMAL}

Options

  --key           Circonus API key/token **${BOLD}REQUIRED${NORMAL}**

  --app           Circonus API app name (authorized w/key) Default: cosi

  [--cosiurl]     COSI URL Default: https://setup.circonus.com/

  [--apiurl]      Circonus API URL Default: https://api.circonus.com/

  [--agent]       Agent mode. Default: reverse
                  reverse = Install circonus-agent, will open connection to broker.
                            broker will request metrics through reverse connection.
                  pull = Install circonus-agent, broker will connect to system and request metrics
                  Note: If circonus-agent is already installed, installation will be skipped

  [--regconf]     Configuration file with custom options to use during registration.

  [--target]      Host IP/hostname to use as check target.

  [--group]       Unique identifier to use when creating/finding the group check. (e.g. 'webservers')

  [--broker]      Broker to use (numeric portion of broker CID e.g. cid=/broker/123, pass 123 as argument).

  [--broker-type] Type of broker to use, (any|enterprise) default: any
                  any - try enterprise brokers, if none available, try public brokers, if none available fail
                  enterprise - only use enterprise brokers, fail if none available

  [--noreg]       Do not attempt to register this system. Using this option
                  will *${BOLD}require${NORMAL}* that the system be manually registered.
                  Default: register system (creating check, graphs, and worksheet)

  [--timeout]     Set timeout for curl operations (default: 120 seconds)

  [--help]        This message

  [--trace]       Enable tracing, output debugging messages
"
}

##
## logging and messaging
##

# ignore tput errors for terms that do not
# support colors (colors will be blank strings)
set +e
RED=$(tput setaf 1)
GREEN=$(tput setaf 2)
NORMAL=$(tput sgr0)
BOLD=$(tput bold)
set -e

log()  { [[ "$cosi_quiet_flag" == 1 ]] && log_only "$*" || printf "%b\n" "$*" | tee -a $cosi_install_log; }
log_only() { printf "%b\n" "${FUNCNAME[1]:-}: $*" >> $cosi_install_log; }
fail() { printf "${RED}" >&2; log "\nERROR: $*\n" >&2; printf "${NORMAL}" >&2; exit 1; }
pass() { printf "${GREEN}"; log "$*"; printf "${NORMAL}"; }

##
## utility functions
##

__parse_parameters() {
    local token=""
    log "Parsing command line parameters"
    while (( $# > 0 )) ; do
        token="$1"
        shift
        case "$token" in
        (--key)
            if [[ -n "${1:-}" ]]; then
                cosi_api_key="$1"
                shift
            else
                fail "--key must be followed by an api key."
            fi
            ;;
        (--app)
            if [[ -n "${1:-}" ]]; then
                cosi_api_app="$1"
                shift
            else
                fail "--app must be followed by an api app."
            fi
            ;;
        (--cosiurl)
            if [[ -n "${1:-}" ]]; then
                cosi_url="$1"
                shift
            else
                fail "--cosiurl must be followed by a URL."
            fi
            ;;
        (--apiurl)
            if [[ -n "${1:-}" ]]; then
                cosi_api_url="$1"
                shift
            else
                fail "--apiurl must be followed by a URL."
            fi
            ;;
        (--agent)
            if [[ -n "${1:-}" ]]; then
                cosi_agent_mode="$1"
                shift
                if [[ ! "${cosi_agent_mode:-}" =~ ^(reverse|pull)$ ]]; then
                    fail "--agent must be followed by a valid agent mode (reverse|pull)."
                fi
            else
                fail "--agent must be followed by an agent mode (reverse|pull)."
            fi
            ;;
        (--regconf)
            if [[ -n "${1:-}" ]]; then
                cosi_regopts_conf="$1"
                shift
            else
                fail "--regconf must be followed by a filespec."
            fi
            ;;
        (--group)
            if [[ -n "${1:-}" ]]; then
                cosi_group_id="$1"
                shift
            else
                fail "--group must be followed by an ID string"
            fi
            ;;
        (--target)
            if [[ -n "${1:-}" ]]; then
                cosi_host_target="$1"
                shift
            else
                fail "--target must be followed by an IP or hostname."
            fi
            ;;
        (--broker)
            if [[ -n "${1:-}" ]]; then
                cosi_broker_id="$1"
                shift
            else
                fail "--broker must be followed by Broker Group ID."
            fi
            ;;
        (--broker-type)
            if [[ -n "${1:-}" ]]; then
                cosi_broker_type="$1"
                shift
                if [[ ! "${cosi_broker_type:-}" =~ ^(any|enterprise)$ ]]; then
                    fail "--broker-type must be followed by type (any|enterprise)."
                fi
            else
                fail "--broker-type must be followed by type (any|enterprise)."
            fi
            ;;
        (--noreg)
            cosi_register_flag=0
            ;;
        (--save)
            cosi_save_config_flag=1
            ;;
        (--trace)
            set -o xtrace
            cosi_trace_flag=1
            ;;
        (--timeout)
            if [[ -n "${1:-}" && "$1" =~ ^\d+$ ]]; then
                cosi_curl_timeout="$1"
                shift
            else
                fail "--timeout must be followed by a value, number of seconds."
            fi
            ;;
        (--help)
            usage
            exit 0
            ;;
        (*)
            printf "\n${RED}Unknown command line option '${token}'.${NORMAL}\n"
            usage
            exit 1
            ;;
        esac
    done
}

__detect_os() {
    local lsb_conf="/etc/lsb-release"
    local release_file=""

    # grab necessary bits of information needed for
    # cosi api to determine if it knows of a agent
    # package to support this system (distro/vers/arch).

    uname -a >> $cosi_install_log

    cosi_os_type="$(uname -s)"
    cosi_os_dist=""
    cosi_os_vers="$(uname -r)"
    cosi_os_arch="$(uname -p)"
    # try 'arch' if 'uname -p' emits 'unknown' (looking at you debian...)
    [[ "$cosi_os_arch" == "unknown" ]] && cosi_os_arch=$(arch)

    set +e
    dmi=$(type -P dmidecode)
    if [[ $? -eq 0 ]]; then
        result=$($dmi -s bios-version 2>/dev/null | tr "\n" " ")
        [[ $? -eq 0 ]] && cosi_os_dmi=$result
    fi
    set -e

    #
    # preference lsb if it is available
    #
    if [[ -f "$lsb_conf" ]]; then
        log "\tLSB found, using '${lsb_conf}' for OS detection."
        cat $lsb_conf >> $cosi_install_log
        source $lsb_conf
        cosi_os_dist="${DISTRIB_ID:-}"
        cosi_os_vers="${DISTRIB_RELEASE:-}"
    fi

    if [[ -z "$cosi_os_dist" ]]; then
        cosi_os_dist="unknown"
        # attempt detection the hard way, thre are way to many methods
        # to "detect" this information and none of them are ubiquitous...
        case "${cosi_os_type}" in
        (Linux)
            if [[ -f /etc/redhat-release ]] ; then
                log "\tAttempt RedHat(variant) detection"
                release_rpm=$(/bin/rpm -qf /etc/redhat-release)
                IFS='-' read -a distro_info <<< "$release_rpm"
                [[ ${#distro_info[@]} -ge 4 ]] || fail "Unable to derive distribution and version from $release_rpm, does not match known pattern."
                case ${distro_info[0]} in
                (centos)
                    # centos-release-5-4.el5.centos.1 - CentOS 5.4
                    # centos-release-6-7.el6.centos.12.3.x86_64 - CentOS 6.7
                    # centos-release-7-2.1511.el7.centos.2.10.x86_64 - CentOS 7.2.1511
                    cosi_os_dist="CentOS"
                    cosi_os_vers="${distro_info[2]}.${distro_info[3]%%\.el*}"
                    ;;
                (redhat)
                    # redhat-release-server-6Server-6.5.0.1.el6.x86_64 - RedHat 6.5.0.1
                    # redhat-release-server-7.2-9.el7.x86_64 - RedHat 7.2
                    cosi_os_dist="RedHat"
                    cosi_os_vers=$(echo $release_rpm | sed -r 's/^.*-([0-9\.]+)(\.el6|-[0-9]).*$/\1/')
                    #[[ ${#distro_info[@]} -ge 5 ]] && cosi_os_vers="${distro_info[4]%%\.el*}"
                    ;;
                (fedora)
                    # fedora-release-23-1.noarch - Fedora 23.1
                    cosi_os_dist="Fedora"
                    cosi_os_vers="${distro_info[2]}.${distro_info[3]%%\.*}"
                    ;;
                (oraclelinux)
                    # oraclelinux-release-7.2-1.0.5.el7.x86_64 - Oracle 7.2
                    cosi_os_dist="Oracle"
                    cosi_os_vers="${distro_info[2]}"
                    ;;
                (*) fail "Unknown RHEL variant '${distro_info[0]}' derived from '${release_rpm}'" ;;
                esac
                log "\tDerived ${cosi_os_dist} v${cosi_os_vers} from '${release_rpm}'"
            elif [[ -f /etc/debian_version ]] ; then
                log "\tAttempt Debian(variant) detection"
                # /etc/debian_version is not consistent enough to be reliable
                # as anything other than a signal. use /etc/os-release or forfeit
                if [[ -f /etc/os-release ]] ; then
                    log_only "\t\tUsing os-release"
                    cat /etc/os-release >> $cosi_install_log
                    source /etc/os-release
                    cosi_os_dist="${ID:-Unsupported}"
                    cosi_os_vers="${VERSION_ID:-}"
                else
                    log_only "\t\tUsing debian_version"
                    cosi_os_dist="Debian"
                    cosi_os_vers="$(head -1 /etc/debian_version)"
                fi
            elif [[ -f /etc/os-release ]]; then
                log "\tAttempt detection from /etc/os-release"
                cat /etc/os-release >> $cosi_install_log
                source /etc/os-release
                if [[ "${PRETTY_NAME:-}" != "" ]]; then
                    log "\t\tFound '${PRETTY_NAME}'"
                fi
                cosi_os_dist="${ID:-Unsupported}"
                cosi_os_vers="${VERSION_ID:-}"
                # if it's an amazon linux ami stuff dmi so it will trigger
                # getting the external public name (amazon linux doesn't
                # include the dmidecode command by default).
                if [[ "${cosi_os_dist:-}" == "amzn" ]]; then
                    if [[ "${cosi_os_dmi:-}" == "" ]]; then
                        cosi_os_dmi="amazon"
                    fi
                fi
            else
                ### add more as needed/supported
                cosi_os_dist="unsup"
            fi
            ;;
        (Darwin)
            cosi_os_dist="OSX"
            ;;
        (FreeBSD|BSD)
            log "\tAttempt ${cosi_os_type} detection"
            if [[ -x /bin/freebsd-version ]]; then
                cosi_os_type="BSD"
                cosi_os_dist="FreeBSD"
                cosi_os_vers=$(/bin/freebsd-version | cut -d '-' -f 1)
            fi
            ;;
        (AIX)
            cosi_os_dist="AIX"
            ;;
        (SunOS|Solaris)
            log "\tAttempt ${cosi_os_type}(variant) detection"
            cosi_os_dist="Solaris"
            cosi_os_arch=$(isainfo -n)
            if [[ -f /etc/release ]]; then
                # dist/version/release signature hopefully on first line...KISSerate
                release_info=$(echo $(head -1 /etc/release))
                log "\tFound /etc/release - using '${release_info}'"
                read -a distro_info <<< "$release_info"
                [[ ${#distro_info[@]} -eq 3 ]] || fail "Unable to derive distribution and version from $release_info, does not match known pattern."
                cosi_os_dist="${distro_info[0]}"
                case "$cosi_os_dist" in
                (OmniOS)
                    cosi_os_vers="${distro_info[2]}"
                    ;;
                (*)
                    cosi_os_vers="${distro_info[1]}.${distro_info[2]}"
                    ;;
                esac
            fi
            ;;
        (*)
            cosi_os_arch="${HOSTTYPE:-}"
            ;;
        esac
    fi
}


__lookup_os() {
    local request_url
    local request_result
    local curl_result
    local cmd_result

    log "\tLooking up $cosi_os_type $cosi_os_dist v$cosi_os_vers $cosi_os_arch."

    #
    # set the global cosi url arguments
    #
    cosi_url_args="?type=${cosi_os_type}&dist=${cosi_os_dist}&vers=${cosi_os_vers}&arch=${cosi_os_arch}"

    request_url="${cosi_url}package/${cosi_url_args}"
    log_only "\tCOSI package request: $request_url"

    #
    # manually handle errors for curl
    #
    set +o errexit

    curl_result=$(\curl -m $cosi_curl_timeout -H 'Accept: text/plain' -sS -w '|%{http_code}' "$request_url" 2>&1)
    cmd_result=$?

    set -o errexit

    log_only "\tResult: \"${curl_result}\" ec=${cmd_result}"

    if [[ $cmd_result -ne 0 ]]; then
        fail "curl command encountered an error (exit code=${cmd_result}) - ${curl_result}\n\tTry curl -v '${request_url}' to see full transaction details."
    fi

    IFS='|' read -a request_result <<< "$curl_result"
    if [[ ${#request_result[@]} -ne 2 ]]; then
        fail "Unexpected response received from COSI request '${curl_result}'. Try curl -v '${request_url}' to see full transaction details."
    fi

    case ${request_result[1]} in
    (200)
        pass "\t$cosi_os_dist $cosi_os_vers $cosi_os_arch supported!"
        IFS='|' read -a cosi_agent_package_info <<< "${request_result[0]//%%/|}"
        ;;
    (000)
        # outlier but, made it happen by trying to get curl to timeout
        # pointed cosi_url at a port being listened to and the daemon responded...doh!
        # (good to know i suppose, that if curl gets a non-http response '000' is the result code)
        fail "Unknown/invalid http result code: ${request_result[1]}\nmessage: ${request_result[0]}"
        ;;
    (*)
        # unsupported distribution|version|architecture
        fail "API result - http result code: ${request_result[1]}\nmessage: ${request_result[0]}"
        ;;
    esac
}


__download_package() {
    local curl_err=""
    local package_url=""
    local package_file=""
    local local_package_file=""

    package_url=${cosi_agent_package_info[0]}
    package_file=${cosi_agent_package_info[1]}

    if [[ "${package_url: -1}" != "/" ]]; then
        package_url+="/"
    fi
    package_url+=$package_file

    #
    # do what we can to validate agent package url
    #
    if [[ -n "${package_url:-}" ]]; then
        [[ "$package_url" =~ ^http[s]?://[^/]+/.*\.(rpm|deb|tgz)$ ]] || fail "COSI agent package url does not match URL pattern (^http[s]?://[^/]+/.*\.(rpm|deb|tgz)$)"
    else
        fail "Invalid COSI agent package url"
    fi

    log "Downloading Agent package ${package_url}"

    local_package_file="${cosi_cache_dir}/${package_file}"

    if [[ -f "${local_package_file}" ]] ; then
        pass "\tFound existing ${local_package_file}, using it for the installation."
    else
        set +o errexit
        \curl -m $cosi_curl_timeout -f "${package_url}" -o "${local_package_file}"
        curl_err=$?
        set -o errexit
        [[ "$curl_err" == "0" && -f "${local_package_file}" ]] || fail "Unable to download '${package_url}' (curl exit code=${curl_err})."
    fi
}

__install_agent() {
    local pub_cmd=""
    local pkg_cmd="${package_install_cmd:-}"
    local pkg_cmd_args="${package_install_args:-}"
    local do_install=""
    local package_file

    if [[ ${cosi_os_type,,} =~ linux ]]; then
        if [[ ${#cosi_agent_package_info[@]} -ne 2 ]]; then
            fail "Invalid Agent package information ${cosi_agent_package_info[@]}, expected 'url file_name'"
        fi
        __download_package
        package_file="${cosi_cache_dir}/${cosi_agent_package_info[1]}"
        log "Installing agent package ${package_file}"
        [[ ! -f "$package_file" ]] && fail "Unable to find package '$package_file'"
        if [[ -z "${pkg_cmd:-}" ]]; then
            if [[ $package_file =~ \.rpm$ ]]; then
                pkg_cmd="yum"
                pkg_cmd_args="localinstall -y ${package_file}"
            elif [[ $package_file =~ \.deb$ ]]; then
                pkg_cmd="dpkg"
                pkg_cmd_args="--install --force-confold ${package_file}"
            else
                fail "Unable to determine package installation command on '${cosi_os_dist}' for '${package_file}'. Please set package_install_cmd in config file to continue."
            fi
        fi
    else
        case "$cosi_os_dist" in
        (OmniOS)
            if [[ ${#cosi_agent_package_info[@]} -ne 3 ]]; then
                fail "Invalid Agent package information ${cosi_agent_package_info[@]}, expected 'publisher_url publisher_name package_name'"
            fi
            set +e
            pkg publisher ${cosi_agent_package_info[1]} &>/dev/null
            if [[ $? -ne 0 ]]; then
                pub_cmd="pkg set-publisher -g ${cosi_agent_package_info[0]} ${cosi_agent_package_info[1]}"
            fi
            set -e
            pkg_cmd="pkg"
            pkg_cmd_args="install ${cosi_agent_package_info[2]}"
            ;;
        (FreeBSD|BSD)
            if [[ ${#cosi_agent_package_info[@]} -ne 2 ]]; then
                fail "Invalid Agent package information ${cosi_agent_package_info[@]}, expected 'url file_name'"
            fi
            __download_package
            package_file="${cosi_cache_dir}/${cosi_agent_package_info[1]}"
            log "Installing agent package ${package_file}"
            [[ ! -f "$package_file" ]] && fail "Unable to find package '$package_file'"
            pkg_cmd="tar"
            pkg_cmd_args="-zxf ${package_file} -C /"
            ;;
        (*)
            fail "Unable to determine package installation command for ${cosi_os_dist}. Please set package_install_cmd in config file to continue."
            ;;
        esac
    fi

    type -P $pkg_cmd >> $cosi_install_log 2>&1 || fail "Unable to find '${pkg_cmd}' command. Ensure it is in the PATH before continuing."

    # callout hook placeholder (PRE)
    if [[ -n "${agent_pre_hook:-}" && -x "${agent_pre_hook}" ]]; then
        log "Agent PRE hook found, running..."
        set +e
        "${agent_pre_hook}"
        set -e
    fi

    if [[ "${pub_cmd:-}" != "" ]]; then
        $pub_cmd 2>&1 | tee -a $cosi_install_log
        [[ ${PIPESTATUS[0]} -eq 0 ]] || fail "adding publisher '${pub_cmd}'"
    fi

    $pkg_cmd $pkg_cmd_args 2>&1 | tee -a $cosi_install_log
    [[ ${PIPESTATUS[0]} -eq 0 ]] || fail "installing ${package_file} '${pkg_cmd} ${pkg_cmd_args}'"

    # reset the agent directory after agent has been installed
    # for the first time.
    [[ -d "${base_dir}/agent" ]] && agent_dir="${base_dir}/agent"
    [[ -d "${agent_dir}/etc" ]] || mkdir -p "${agent_dir}/etc"

    # callout hook placeholder (POST)
    if [[ -n "${agent_post_hook:-}" && -x "${agent_post_hook}" ]]; then
        log "Agent POST hook found, running..."
        set +e
        "${agent_post_hook}"
        set -e
    fi

    # give agent a couple seconds to start/restart
    sleep 2
}

__is_agent_installed() {
    local agent_bin="${agent_dir}/sbin/circonus-agentd"
    if [[ -x "$agent_bin" ]]; then
        pass "agent installation found"
        set +e
        if [[ -x /usr/bin/dpkg-query ]]; then
            log_only "\t$(/usr/bin/dpkg-query --show circonus-agent 2>&1)"
            agent_pkg_ver=$(/usr/bin/dpkg-query --showformat='${Version}' --show circonus-agent)
            [[ $? -ne 0 ]] && agent_pkg_ver=""
        elif [[ -x /usr/bin/rpm ]]; then
            log_only "\t$(/usr/bin/rpm -qi circonus-agent 2>&1)"
            agent_pkg_ver=$(/usr/bin/rpm --queryformat '%{Version}' -q circonus-agent 2>/dev/null)
            [[ $? -ne 0 ]] && agent_pkg_ver=""
        elif [[ -x /usr/bin/pkg ]]; then
            log_only "\t$(/usr/bin/pkg info field/circonus-agent 2>&1)"
        else
            log_only "\tagent found but do not know how to get info for this OS."
        fi
        set -e
        agent_state=1
    fi
}

__is_agent_running() {
    local pid
    local ret
    if [[ $agent_state -eq 1 ]]; then
        set +e
        pid=$(pgrep -n -f "sbin/circonus-agentd")
        ret=$?
        set -e
        if [[ $ret -eq 0 && ${pid:-0} -gt 0 ]]; then
            pass "agent process running PID:${pid}"
            agent_state=2
        fi
    fi
}

__check_agent_url() {
    local url="${agent_url:-http://127.0.0.1:2609/}inventory"
    local err
    local ret

    if [[ $agent_state -eq 2 ]]; then
        set +e
        err=$(\curl --noproxy localhost,127.0.0.1 -sSf "$url" -o /dev/null 2>&1)
        ret=$?
        set -e
        if [[ $ret -ne 0 ]]; then
            fail "agent installed and running but not reachable\nCurl exit code: ${ret}\nCurl err msg: ${err}"
        fi
        pass "agent URL reachable"
        agent_state=3
    fi
}

__check_agent() {
    if [[ $agent_state -eq 0 ]]; then
        __is_agent_installed  #state 1
    fi
    if [[ $agent_state -eq 1 ]]; then
        __is_agent_running    #state 2
    fi
    if [[ $agent_state -eq 2 ]]; then
        __check_agent_url     #state 3
    fi
}

__start_agent() {
    local agent_pid
    local ret

    log "Starting agent (if not already running)"

    if [[ ${agent_state:-0} -eq 1 ]]; then
        if [[ -s /lib/systemd/system/circonus-agent.service ]]; then
            systemctl start circonus-agent
        elif [[ -s /etc/init/circonus-agent.conf ]]; then
            initctl start circonus-agent
        elif [[ -s /etc/init.d/circonus-agent ]]; then
            /etc/init.d/circonus-agent start
        elif [[ -s /var/svc/manifest/network/circonus/circonus-agent.xml ]]; then
            svcadm enable circonus-agent
        elif [[ -s /etc/rc.d/circonus-agent ]]; then
            # create blank /etc/rc.conf if it does not exist
            # e.g. an outlier use case: bare install which has not had
            # even the most rudimentary operational configuration tasks
            # perfomed but, cosi is being run on it for whatever reason...
            [[ -f /etc/rc.conf ]] || touch /etc/rc.conf
            if [[ -f /etc/rc.conf ]]; then
                # treat as FreeBSD
                # enable it if there is no circonus-agent_enable setting
                [[ $(grep -cE '^circonus_agent_enable' /etc/rc.conf) -eq 0 ]] && echo 'circonus_agent_enable="YES"' >> /etc/rc.conf
                # start it if it is enabled
                [[ $(grep -c 'circonus_agent_enable="YES"' /etc/rc.conf) -eq 1 ]] && service circonus-agent start
            fi
        else
            fail "Agent installed, unable to determine how to start it (unrecognized init system)."
        fi
        if [[ $? -ne 0 ]]; then
            fail "COSI was unable to start circonus-agent - try running circonus-agent manually, check for errors specific to this system."
        fi
        sleep 5
    fi

    set +e
    agent_pid=$(pgrep -n -f "sbin/circonus-agentd")
    ret=$?
    set -e

    if [[ ${ret:-0} -eq 0 && ${agent_pid:-0} -gt 0 ]]; then
        pass "Agent running with PID ${agent_pid}"
    else
        log "'pgrep -n -f \"sbin/circonus-agentd\"' exited with code (${ret})."
        if [[ "${cosi_os_type}" == "Linux" ]]; then
            if [[ ${ret:-0} -eq 1 ]]; then
                log "Exit code ${ret} means more than one process was found."
            elif [[ ${ret:-0} -eq 2 ]]; then
                log "Exit code ${ret} means no process matched, after attempt to start."
            fi
        fi
        fail "Unable to locate running agent, check agent log or run interactively '/opt/circonus/agent/sbin/circonus-agentd' for more information."
    fi
}

__restart_agent() {
    if [[ -s /lib/systemd/system/circonus-agent.service ]]; then
        systemctl restart circonus-agent
    elif [[ -s /etc/init/circonus-agent.conf ]]; then
        initctl restart circonus-agent
    elif [[ -s /etc/init.d/circonus-agent ]]; then
        /etc/init.d/circonus-agent restart
    elif [[ -s /var/svc/manifest/network/circonus/circonus-agent.xml ]]; then
        svcadm restart circonus-agent
    elif [[ -s /etc/rc.d/circonus-agent ]]; then
        if [[ -f /etc/rc.conf ]]; then
            [[ $(grep -c 'circonus-agent_enable="YES"' /etc/rc.conf) -eq 1 ]] && service circonus-agent restart
        fi
    else
        fail "Agent installed, unable to determine how to restart it (unrecognized init system)."
    fi
}

__save_cosi_register_config() {
    #
    # saves the cosi-install configuraiton options for cosi-register
    #
    log "Saving COSI registration configuration ${cosi_register_config}"
    cat <<EOF > "$cosi_register_config"
---
# generated by cosi script on $(date)
cosi_url: "${cosi_url}"
base_ui_url: ""
debug: false
agent:
    mode: "${cosi_agent_mode}"
    url: "${agent_url}"
api:
    key: "${cosi_api_key}"
    app: "${cosi_api_app}"
    url: "${cosi_api_url}"
    ca_file: ""
checks:
  target: "${cosi_host_target}"
  group_id: ${cosi_group_id}
  broker:
    id: ${cosi_broker_id}
    type: ${cosi_broker_type}
reg_conf: "${cosi_regopts_conf}"
log:
    level: info
    pretty: true
system:
    arch: "${cosi_os_arch}"
    os_type: "${cosi_os_type}"
    os_dist: "${cosi_os_dist}"
    os_vers: "${cosi_os_vers}"
    dmi: "${cosi_os_dmi}"
EOF
    [[ $? -eq 0 ]] || fail "Unable to save COSI registration configuration '${cosi_register_config}'"
    [[ -f ${cosi_register_id_file} ]] || echo $cosi_id > ${cosi_register_id_file}
}


__fetch_cosi_tool() {
    local cosi_url_args="?type=${cosi_os_type}&dist=${cosi_os_dist}&vers=${cosi_os_vers}&arch=${cosi_os_arch}"
    local cosi_tool_url="${cosi_url}tool/"
    local cosi_tool_file="${cosi_cache_dir}/cosi-tool.tgz"
    local curl_err

    log "Retrieving COSI tool ${cosi_tool_url}"
    log_only "\tTool URL : $cosi_tool_url"
    log_only "\tTool file: $cosi_tool_file"

    set +o errexit
    \curl -m $cosi_curl_timeout -f -sSL "${cosi_tool_url}${cosi_url_args}" -o "${cosi_tool_file}"
    curl_err=$?
    set -o errexit
    [[ $curl_err -eq 0 && -f "$cosi_tool_file" ]] || {
        [[ -f "$cosi_tool_file" ]] && rm "$cosi_tool_file"
        fail "Unable to fetch '${cosi_tool_url}' (curl exit code=${curl_err})."
    }

    cd "$cosi_dir"
    log "Unpacking COSI utilities into $(pwd)"
    tar -oxf "$cosi_tool_file"
    [[ $? -eq 0 ]] || fail "Unable to unpack COSI tool"
}


#
# main support functions
#

cosi_initialize() {
    local settings_list

    #
    # precedence order:
    #   load config (if exists)
    #   backfill with defaults
    #   override with command line (vars/flags)
    #

    log "Initializing cosi-install"

    if [[ "$*" == *--trace* ]] || (( ${cosi_trace_flag:-0} > 0 )) ; then
      set -o xtrace
      cosi_trace_flag=1
    fi

    BASH_MIN_VERSION="3.2.25"
    if [[ -n "${BASH_VERSION:-}" &&
          "$(printf "%b" "${BASH_VERSION:-}\n${BASH_MIN_VERSION}\n" | LC_ALL=C sort -t"." -k1,1n -k2,2n -k3,3n | head -n1)" != "${BASH_MIN_VERSION}" ]]; then
        fail "BASH ${BASH_MIN_VERSION} required (you have $BASH_VERSION)"
    fi

    export PS4="+ \${FUNCNAME[0]:+\${FUNCNAME[0]}()}  \${LINENO} > "

    #
    # enable use of a config file for automated deployment support
    #
    if [[ -f "$cosi_config_file" ]] ; then
        log_only "Loading config file ${cosi_config_file}"
        source "$cosi_config_file"
    fi

    #
    # internal variables (after sourcing config file, prevent unintentional overrides)
    #
    base_dir="/opt/circonus"
    agent_dir="${base_dir}"
    [[ -d "${base_dir}/agent" ]] && agent_dir="${base_dir}/agent"

    bin_dir="${base_dir}/bin"
    etc_dir="${cosi_dir}/etc"
    reg_dir="${cosi_dir}/registration"

    agent_state=0
    agent_ip="127.0.0.1"
    agent_port="2609"
    agent_url="http://${agent_ip}:${agent_port}/"
    agent_pre_hook="${cosi_dir}/agent_pre_hook.sh"
    agent_post_hook="${cosi_dir}/agent_post_hook.sh"
    cosi_agent_package_info=()
    cosi_cache_dir="${cosi_dir}/cache"
    cosi_register_config="${etc_dir}/cosi.yaml"
    cosi_register_id_file="${etc_dir}/.cosi_id"
    cosi_url_args=""
    cosi_util_dir="${cosi_dir}/util"
    cosi_os_arch=""
    cosi_os_dist=""
    cosi_os_type=""
    cosi_os_vers=""
    cosi_os_dmi=""
    cosi_id=""
    cosi_group_id=""

    #
    # set defaults (if config file not used or options left unset)
    #
    : ${cosi_trace_flag:=0}
    : ${cosi_quiet_flag:=0}
    : ${cosi_register_flag:=1}
    : ${cosi_regopts_conf:=}
    : ${cosi_host_target:=}
    : ${cosi_broker_id:=}
    : ${cosi_broker_type:=any}
    : ${cosi_save_config_flag:=0}
    : ${cosi_url:=https://setup.circonus.com/}
    : ${cosi_api_url:=https://api.circonus.com/}
    : ${cosi_api_key:=}
    : ${cosi_api_app:=cosi}
    : ${cosi_agent_mode:=reverse}
    : ${cosi_install_agent:=1}
    : ${package_install_cmd:=}
    : ${package_install_args:=--install}
    : ${cosi_curl_timeout:=120}

    # list of settings we will save if cosi_save_config_flag is ON
    settings_list=" \
    cosi_api_key \
    cosi_api_app \
    cosi_api_url \
    cosi_url \
    cosi_agent_mode \
    cosi_host_target \
    cosi_broker_id \
    cosi_broker_type \
    cosi_install_agent \
    cosi_register_flag \
    cosi_register_config \
    cosi_quiet_flag \
    package_install_cmd \
    package_install_args \
    cosi_group_id
    "

    #
    # manually handle errors for these
    #
    set +o errexit

    # let environment VARs override config/default for api key/app
    [[ -n "${COSI_KEY:-}" ]] && cosi_api_key="$COSI_KEY"
    [[ -n "${COSI_APP:-}" ]] && cosi_api_app="$COSI_APP"

    #
    # trigger error if needed commands are not found...
    local cmd_list="awk cat chgrp chmod curl grep head ln mkdir pgrep sed tar tee uname"
    local cmd
    log_only "Verifying required commands exist. '${cmd_list}'"
    for cmd in $cmd_list; do
        type -P $cmd >> $cosi_install_log 2>&1 || fail "Unable to find '${cmd}' command. Ensure it is available in PATH '${PATH}' before continuing."
    done

    set -o errexit

    #
    # parameters override defaults and config file settings (if it was used)
    #
    __parse_parameters "$@"

    #
    # verify *required* values API key and app
    #
    [[ -n "${cosi_api_key:-}" ]] || fail "API key is *required*. (see '${0} -h' for more information.)"
    [[ -n "${cosi_api_app:-}" ]] || fail "API app is *required*. (see '${0} -h' for more information.)"

    #
    # fixup URLs, ensure they end with '/'
    #
    [[ "${cosi_api_url: -1}" == "/" ]] || cosi_api_url+="/"
    [[ "${cosi_url: -1}" == "/" ]] || cosi_url+="/"
    [[ "${agent_url: -1}" == "/" ]] || agent_url+="/"

    type -P uuidgen > /dev/null 2>&1 && cosi_id=$(uuidgen)

    if [[ -z "${cosi_id:-}" ]]; then
        kern_uuid=/proc/sys/kernel/random/uuid
        if [[ -f $kern_uuid ]]; then
            cosi_id=$(cat $kern_uuid)
        else
            cosi_id=$(python  -c 'import uuid; print uuid.uuid1()')
        fi
    fi

    hn=$(hostname)
    if [[ -z "${hn:-}" && -z "${cosi_host_target:-}" ]]; then
        fatal "cosi requires system hostname to be set or overridden with --target"
    fi

    #
    # optionally, save the cosi-install config
    # (can be used on other systems and/or during testing)
    #
    if [[ "${cosi_save_config_flag:-0}" == "1" ]] ; then
        log "Saving config file ${cosi_config_file}"
        > "$cosi_config_file"
        for cosi_setting in $settings_list; do
            echo "${cosi_setting}=\"${!cosi_setting}\"" >> "$cosi_config_file"
        done
    fi

    [[ -d "$cosi_cache_dir" ]] || {
        mkdir -p "$cosi_cache_dir"
        [[ $? -eq 0 ]] || fail "Unable to create cache_dir '${cosi_cache_dir}'."
    }
    [[ -d "$reg_dir" ]] || {
        mkdir -p "$reg_dir"
        [[ $? -eq 0 ]] || fail "Unable to create reg_dir '${reg_dir}'."
    }
    [[ -d "$etc_dir" ]] || {
        mkdir -p "$etc_dir"
        [[ $? -eq 0 ]] || fail "Unable to create etc_dir '${etc_dir}'."
    }
    [[ -d "$bin_dir" ]] || {
        mkdir -p "$bin_dir"
        [[ $? -eq 0 ]] || fail "Unable to create bin_dir '${bin_dir}'."
    }

}


cosi_verify_os() {
    log "Verifying COSI support for OS"
    __detect_os
    __lookup_os
}


cosi_check_agent() {
    log "Checking Agent state"
    __check_agent

    if [[ $agent_state -eq 0 ]]; then
        log "Agent not found, installing Agent"
        __install_agent
        __check_agent
    fi

    if [[ $agent_state -eq 1 ]]; then
        __start_agent
        __check_agent
    fi

    if [[ $agent_state -eq 3 ]]; then
        pass "Agent running and responding"
    fi
}


cosi_register() {
    local cosi_script="${cosi_dir}/bin/cosi"
    local cosi_register_cmd="register"
    local cosi_register_opt=""
    local install_reverse="${cosi_dir}/bin/reverse_install.sh"
    local agent_config="$agent_dir/etc/circonus-agent.yaml"
    local agent_bin="${agent_dir}/sbin/circonus-agentd"
    local cosi_host="${cosi_host_target:-$(hostname)}"

    echo
    __fetch_cosi_tool
    echo
    __save_cosi_register_config
    echo

    if [[ "${cosi_register_flag:-1}" != "1" ]]; then
        log "Not running COSI registration script, --noreg requested"
        return
    fi

    [[ -x "$cosi_script" ]] || fail "Unable to find cosi command '${cosi_script}'"

    echo
    log "### Creating base circonus-agent configuration"
    echo
    $agent_bin --check-metric-streamtags --check-tags="host:${cosi_host},os:${cosi_os_type},arch:${cosi_os_arch},distro:${cosi_os_dist}-${cosi_os_vers}" --show-config=yaml > $agent_config
    [[ $? -eq 0 ]] || fail "Error creating base circonus-agent configuration"
    __restart_agent

    echo
    log "### Running COSI registration ###"
    echo
    log_only "running: $cosi_script" "$cosi_register_cmd"
    "$cosi_script" "$cosi_register_cmd" | tee -a $cosi_install_log
    [[ ${PIPESTATUS[0]} -eq 0 ]] || fail "Errors encountered during registration."

    if [[ "${cosi_agent_mode}" == "reverse" ]]; then
        echo
        log "### Enabling ${cosi_agent_mode} mode for agent ###"
        echo
        $agent_bin --listen=127.0.0.1:2609 --reverse --check-id=cosi --api-key=cosi --api-app=cosi --show-config=yaml --check-metric-streamtags --check-tags="host:${cosi_host},os:${cosi_os_type},arch:${cosi_os_arch},distro:${cosi_os_dist}-${cosi_os_vers}" > $agent_config
        [[ $? -eq 0 ]] || fail "Error updating circonus-agent configuration"
        __restart_agent
    fi

    echo
    log "### Registration Overview ###"
    echo
    pass "--- Graphs created ---"
    log "running: '${cosi_dir}/bin/cosi graph list --long'"
    "${cosi_dir}/bin/cosi" graph list --long

    echo
    pass "--- Check created ---"
    log "running: '${cosi_dir}/bin/cosi check list --long --verify'"
    "${cosi_dir}/bin/cosi" check list --long --verify

    echo
    pass "--- Worksheet created ---"
    log "running: '${cosi_dir}/bin/cosi worksheet list --long'"
    "${cosi_dir}/bin/cosi" worksheet list --long

    echo
    pass "--- Dashboard created ---"
    log "running: '${cosi_dir}/bin/cosi dashboard list --long'"
    "${cosi_dir}/bin/cosi" dashboard list --long

    echo
    echo "To see any of these lists again in the future run, ${cosi_dir}/bin/cosi (graph|check|worksheet|dashboard) list --long"
    echo
}


cosi_install() {
    cosi_initialize "$@"
    #
    # short circuit if NAD installation detected
    #
    if [[ -d "${base_dir}/nad" || -f "${base_dir}/sbin/nad.js" ]]; then
        echo
        echo
        log "*** Previous NAD installation detected ***"
        log "Please use the following crconus-agent installation instructions:"
        log "https://github.com/circonus-labs/circonus-agent#quick-start"
        echo
        echo
        exit
    fi
    cosi_verify_os
    cosi_check_agent
    cosi_register
}

####
################### main
####

#
# short-circuit a request for help
#
if [[ "$*" == *--help* ]]; then
    usage
    exit 0
fi

#
# no arguments are passed and no conf file
#
if [[ $# -eq 0 && ! -f "$cosi_config_file" ]]; then
    usage
    exit 0
fi


#
# NOTE Ensure sufficient rights to do the install
#
(( UID != 0 )) && {
    printf "\n%b\n\n" "${RED}Must run as root[sudo] -- installing software requires certain permissions.${NORMAL}"
    exit 1
}

#
# NOTE All COSI assets and logs are saved in the cosi_dir
#
: ${cosi_dir:=/opt/circonus/cosi}
[[ -d "$cosi_dir" ]] || {
    set +e
    mkdir -p "$cosi_dir"
    [[ $? -eq 0 ]] || {
        printf "\n%b\n" "${RED}Unable to create cosi_dir '${cosi_dir}'.${NORMAL}"
        exit 1
    }
    set -e
}
cosi_log_dir="${cosi_dir}/log"
[[ -d "$cosi_log_dir" ]] || {
    set +e
    mkdir -p "$cosi_log_dir"
    [[ $? -eq 0 ]] || {
        printf "\n%b\n" "${RED}Unable to create cosi_log_dir '${cosi_log_dir}'.${NORMAL}"
        exit 1
    }
    set -e
}
cosi_install_log="${cosi_log_dir}/install.log"

#
# squelch output (log messages to file only)
#
: ${cosi_quiet_flag:=0}
if [[ "$*" == *--quiet* ]]; then
  cosi_quiet_flag=1
fi

[[ ! -f $cosi_install_log ]] || printf "\n\n==========\n\n" >> $cosi_install_log
log "Started Circonus One step Install on $(date)"
printf "Options: $*\n" >> $cosi_install_log

cosi_install "$@"

log "Completed Circonus One step Install on $(date)\n"

## END
# vim:ts=4:sw=4:et
