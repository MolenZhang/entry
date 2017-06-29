SET(CPACK_GENERATOR "RPM")
SET(CPACK_PACKAGE_NAME "entry")
SET(CPACK_PACKAGE_VERSION "1.0")
SET(CPACK_PACKAGE_RELEASE "0")
SET(CPACK_PACKAGE_DESCRIPTION "entry which will be installed in /opt/")
SET(CPACK_PACKAGE_FILE_NAME
"${CPACK_PACKAGE_NAME}-${CPACK_PACKAGE_VERSION}-${CPACK_PACKAGE_RELEASE}.${CMAKE_SYSTEM_PROCESSOR}")
SET(CPACK_INSTALL_COMMANDS
#"rm -rf $ENV{PWD}/build"
#"rm -rf $ENV{PWD}/build/svc"
#"rm -rf $ENV{PWD}/build/cfg"
#"rm -rf $ENV{PWD}/build/bin"
"mkdir -p $ENV{PWD}/build/"
"mkdir -p $ENV{PWD}/build/etc/sysconfig/entry/"
"mkdir -p $ENV{PWD}/build/usr/bin/"
"mkdir -p $ENV{PWD}/build/usr/lib/systemd/system/"
#"cp $ENV{PWD}/nodeIP.cfg /$ENV{PWD}/build/etc/sysconfig/k8sNginxEX/"
"cp $ENV{PWD}/entry.cfg /$ENV{PWD}/build/etc/sysconfig/entry/"
"cp $ENV{PWD}/entry /$ENV{PWD}/build/usr/bin/"
"chmod 777 /$ENV{PWD}/build/usr/bin/entry"
"cp $ENV{PWD}/entry.service /$ENV{PWD}/build/usr/lib/systemd/system/"
)
SET(CPACK_INSTALLED_DIRECTORIES
#SET(CPACK_PACKAGE_INSTALL_DIRECTORY
"$ENV{PWD}/build;/")

