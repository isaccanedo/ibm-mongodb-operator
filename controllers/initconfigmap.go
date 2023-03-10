//
// Copyright 2021 IBM Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
package controllers

const initConfigMap = `
---
# Source: icp-mongodb/templates/mongodb-init-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  labels:
    app.kubernetes.io/name: icp-mongodb
    app.kubernetes.io/instance: icp-mongodb
    app.kubernetes.io/version: 4.0.12-build.3
    app.kubernetes.io/component: database
    app.kubernetes.io/part-of: common-services-cloud-pak
    app.kubernetes.io/managed-by: operator
    release: mongodb
  name: icp-mongodb-init
data:
  on-start.sh: |
    #!/bin/bash

    ## workaround https://serverfault.com/questions/713325/openshift-unable-to-write-random-state
    export RANDFILE=/tmp/.rnd
    port=27017
    replica_set=$REPLICA_SET
    script_name=${0##*/}
    credentials_file=/work-dir/credentials.txt
    config_dir=/data/configdb

    function log() {
        local msg="$1"
        local timestamp=$(date --iso-8601=ns)
        1>&2 echo "[$timestamp] [$script_name] $msg"
        echo "[$timestamp] [$script_name] $msg" >> /work-dir/log.txt
    }

    if [[ "$AUTH" == "true" ]]; then

        if [ !  -f "$credentials_file" ]; then
            log "Creds File Not found!"
            log "Original User: $ADMIN_USER"
            echo $ADMIN_USER > $credentials_file
            echo $ADMIN_PASSWORD >> $credentials_file
        fi
        admin_user=$(head -n 1 $credentials_file)
        admin_password=$(tail -n 1 $credentials_file)
        admin_auth=(-u "$admin_user" -p "$admin_password")
        log "Original User: $admin_user"
        if [[ "$METRICS" == "true" ]]; then
            metrics_user="$METRICS_USER"
            metrics_password="$METRICS_PASSWORD"
        fi
    fi

    function shutdown_mongo() {

        log "Running fsync..."
        mongo admin "${admin_auth[@]}" "${ssl_args[@]}" --eval "db.adminCommand( { fsync: 1, lock: true } )"

        log "Running fsync unlock..."
        mongo admin "${admin_auth[@]}" "${ssl_args[@]}" --eval "db.adminCommand( { fsyncUnlock: 1 } )"

        log "Shutting down MongoDB..."
        mongo admin "${admin_auth[@]}" "${ssl_args[@]}" --eval "db.adminCommand({ shutdown: 1, force: true, timeoutSecs: 60 })"
    }

    #Check if Password has change and updated in mongo , if so update Creds
    function update_creds_if_changed() {
      if [ "$admin_password" != "$ADMIN_PASSWORD" ]; then
          passwd_changed=true
          log "password has changed = $passwd_changed"
          log "checking if passwd  updated in mongo"
          mongo admin  "${ssl_args[@]}" --eval "db.auth({user: '$admin_user', pwd: '$ADMIN_PASSWORD'})" | grep "Authentication failed"
          if [[ $? -eq 1 ]]; then
            log "New Password worked, update creds"
            echo $ADMIN_USER > $credentials_file
            echo $ADMIN_PASSWORD >> $credentials_file
            admin_password=$ADMIN_PASSWORD
            admin_auth=(-u "$admin_user" -p "$admin_password")
            passwd_updated=true
          fi
      fi
    }

    function update_mongo_password_if_changed() {
      log "checking if mongo passwd needs to be  updated"
      if [[ "$passwd_changed" == "true" ]] && [[ "$passwd_updated" != "true" ]]; then
        log "Updating to new password "
        if [[ $# -eq 1 ]]; then
            mhost="--host $1"
        else
            mhost=""
        fi

        log "host for password upd ($mhost)"
        mongo admin $mhost "${admin_auth[@]}" "${ssl_args[@]}" --eval "db.changeUserPassword('$admin_user', '$ADMIN_PASSWORD')" >> /work-dir/log.txt 2>&1
        sleep 10
        log "mongo passwd change attempted; check and update creds file if successful"
        update_creds_if_changed
      fi
    }



    my_hostname=$(hostname)
    log "Bootstrapping MongoDB replica set member: $my_hostname"

    log "Reading standard input..."
    while read -ra line; do
        log "line is  ${line}"
        if [[ "${line}" == *"${my_hostname}"* ]]; then
            service_name="$line"
        fi
        peers=("${peers[@]}" "$line")
    done

    # Move into /work-dir
    pushd /work-dir
    pwd >> /work-dir/log.txt
    ls -l  >> /work-dir/log.txt

    # Generate the ca cert
    ca_crt=$config_dir/tls.crt
    if [ -f $ca_crt  ]; then
        log "Generating certificate"
        ca_key=$config_dir/tls.key
        pem=/work-dir/mongo.pem
        ssl_args=(--ssl --sslCAFile $ca_crt --sslPEMKeyFile $pem)

        echo "ca stuff created" >> /work-dir/log.txt

    cat >openssl.cnf <<EOL
    [req]
    req_extensions = v3_req
    distinguished_name = req_distinguished_name
    [req_distinguished_name]
    [ v3_req ]
    basicConstraints = CA:FALSE
    keyUsage = nonRepudiation, digitalSignature, keyEncipherment
    subjectAltName = @alt_names
    [alt_names]
    DNS.1 = $(echo -n "$my_hostname" | sed s/-[0-9]*$//)
    DNS.2 = $my_hostname
    DNS.3 = $service_name
    DNS.4 = localhost
    DNS.5 = 127.0.0.1
    DNS.6 = mongodb
    EOL

        # Generate the certs
        echo "cnf stuff" >> /work-dir/log.txt
        echo "genrsa " >> /work-dir/log.txt
        openssl genrsa -out mongo.key 2048 >> /work-dir/log.txt 2>&1

        echo "req " >> /work-dir/log.txt
        openssl req -new -key mongo.key -out mongo.csr -subj "/CN=$my_hostname" -config openssl.cnf >> /work-dir/log.txt 2>&1

        echo "x509 " >> /work-dir/log.txt
        openssl x509 -req -in mongo.csr \
            -CA $ca_crt -CAkey $ca_key -CAcreateserial \
            -out mongo.crt -days 3650 -extensions v3_req -extfile openssl.cnf >> /work-dir/log.txt 2>&1

        echo "mongo stuff" >> /work-dir/log.txt

        rm mongo.csr

        cat mongo.crt mongo.key > $pem
        rm mongo.key mongo.crt
    fi


    log "Peers: ${peers[@]}"

    log "Starting a MongoDB instance..."
    mongod --config $config_dir/mongod.conf >> /work-dir/log.txt 2>&1 &
    pid=$!
    trap shutdown_mongo EXIT


    log "Waiting for MongoDB to be ready..."
    until [[ $(mongo "${ssl_args[@]}" --quiet --eval "db.adminCommand('ping').ok") == "1" ]]; do
        log "Retrying..."
        sleep 2
    done

    log "Initialized."

    if [[ "$AUTH" == "true" ]]; then
        update_creds_if_changed
    fi

    iter_counter=0
    while [  $iter_counter -lt 5 ]; do
      log "primary check, iter_counter is $iter_counter"
      # try to find a master and add yourself to its replica set.
      for peer in "${peers[@]}"; do
          log "Checking if ${peer} is primary"
          mongo admin --host "${peer}" --ipv6 "${admin_auth[@]}" "${ssl_args[@]}" --quiet --eval "rs.status()"  >> log.txt

          # Check rs.status() first since it could be in primary catch up mode which db.isMaster() doesn't show
          if [[ $(mongo admin --host "${peer}" --ipv6 "${admin_auth[@]}" "${ssl_args[@]}" --quiet --eval "rs.status().myState") == "1" ]]; then
              log "Found master ${peer}, wait while its in primary catch up mode "
              until [[ $(mongo admin --host "${peer}" --ipv6 "${admin_auth[@]}" "${ssl_args[@]}" --quiet --eval "db.isMaster().ismaster") == "true" ]]; do
                  sleep 1
              done
              primary="${peer}"
              log "Found primary: ${primary}"
              break
          fi
      done

      if [[ -z "${primary}" ]]  && [[ ${#peers[@]} -gt 1 ]] && (mongo "${ssl_args[@]}" --eval "rs.status()" | grep "no replset config has been received"); then
        log "waiting before creating a new replicaset, to avoid conflicts with other replicas"
        sleep 30
      else
        break
      fi

      let iter_counter=iter_counter+1
    done


    if [[ "${primary}" = "${service_name}" ]]; then
        log "This replica is already PRIMARY"

    elif [[ -n "${primary}" ]]; then

        if [[ $(mongo admin --host "${primary}" --ipv6 "${admin_auth[@]}" "${ssl_args[@]}" --quiet --eval "rs.conf().members.findIndex(m => m.host == '${service_name}:${port}')") == "-1" ]]; then
          log "Adding myself (${service_name}) to replica set..."
          if (mongo admin --host "${primary}" --ipv6 "${admin_auth[@]}" "${ssl_args[@]}" --eval "rs.add('${service_name}')" | grep 'Quorum check failed'); then
              log 'Quorum check failed, unable to join replicaset. Exiting.'
              exit 1
          fi
        fi
        log "Done,  Added myself to replica set."

        sleep 3
        log 'Waiting for replica to reach SECONDARY state...'
        until printf '.'  && [[ $(mongo admin "${admin_auth[@]}" "${ssl_args[@]}" --quiet --eval "rs.status().myState") == '2' ]]; do
            sleep 1
        done
        log '??? Replica reached SECONDARY state.'

    elif (mongo "${ssl_args[@]}" --eval "rs.status()" | grep "no replset config has been received"); then

        log "Initiating a new replica set with myself ($service_name)..."

        mongo "${ssl_args[@]}" --eval "rs.initiate({'_id': '$replica_set', 'members': [{'_id': 0, 'host': '$service_name'}]})"
        mongo "${ssl_args[@]}" --eval "rs.status()"

        sleep 3

        log 'Waiting for replica to reach PRIMARY state...'

        log ' Waiting for rs.status state to become 1'
        until printf '.'  && [[ $(mongo "${ssl_args[@]}" --quiet --eval "rs.status().myState") == '1' ]]; do
            sleep 1
        done

        log ' Waiting for master to complete primary catchup mode'
        until [[ $(mongo  "${ssl_args[@]}" --quiet --eval "db.isMaster().ismaster") == "true" ]]; do
            sleep 1
        done

        primary="${service_name}"
        log '??? Replica reached PRIMARY state.'


        if [[ "$AUTH" == "true" ]]; then
            # sleep a little while just to be sure the initiation of the replica set has fully
            # finished and we can create the user
            sleep 3

            log "Creating admin user..."
            mongo admin "${ssl_args[@]}" --eval "db.createUser({user: '$admin_user', pwd: '$admin_password', roles: [{role: 'root', db: 'admin'}]})"
        fi

        log "Done initiating replicaset."

    fi

    log "Primary: ${primary}"

    if [[  -n "${primary}"   && "$AUTH" == "true" ]]; then
        # you r master and passwd has changed.. then update passwd
        update_mongo_password_if_changed $primary

        if [[ "$METRICS" == "true" ]]; then
            log "Checking if metrics user is already created ..."
            metric_user_count=$(mongo admin --host "${primary}" "${admin_auth[@]}" "${ssl_args[@]}" --eval "db.system.users.find({user: '${metrics_user}'}).count()" --quiet)
            log "User count is ${metric_user_count} "
            if [[ "${metric_user_count}" == "0" ]]; then
                log "Creating clusterMonitor user... user - ${metrics_user}  "
                mongo admin --host "${primary}" "${admin_auth[@]}" "${ssl_args[@]}" --eval "db.createUser({user: '${metrics_user}', pwd: '${metrics_password}', roles: [{role: 'clusterMonitor', db: 'admin'}, {role: 'read', db: 'local'}]})"
                log "User creation return code is $? "
                metric_user_count=$(mongo admin --host "${primary}" "${admin_auth[@]}" "${ssl_args[@]}" --eval "db.system.users.find({user: '${metrics_user}'}).count()" --quiet)
                log "User count now is ${metric_user_count} "
            fi
        fi
    fi

    log "MongoDB bootstrap complete"
    exit 0`
