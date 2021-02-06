with import <nixpkgs> {};

stdenv.mkDerivation {
    name = "postgresql-inside-nixshell";

    buildInputs = [
        glibcLocales
        (pkgs.postgresql.withPackages (p: [ p.postgis ]))
        sqitchPg
        pgcli
    ];

    shellHook = ''
        export PGDATA="$PWD/db"
        export SOCKET_DIRECTORIES="$PWD/sockets"
        mkdir $SOCKET_DIRECTORIES
        initdb
        echo "unix_socket_directories = '$SOCKET_DIRECTORIES'" >> $PGDATA/postgresql.conf
        pg_ctl -l $PGDATA/logfile start
        createuser postgres --createdb -h localhost
        function end {
          echo "Shutting down the database..."
          pg_ctl stop
          echo "Removing directories..."
          rm -rf $PGDATA $SOCKET_DIRECTORIES
        }
        trap end EXIT
    '';
}
