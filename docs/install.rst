Install teragrid
==================

From Binary
-----------

To download pre-built binaries, see the `Download page <https://teragrid.com/downloads>`__.

From Source
-----------

You'll need ``go``, maybe `dep <https://github.com/golang/dep>`__, and the teragrid source code.

Install Go
^^^^^^^^^^

Make sure you have `installed Go <https://golang.org/doc/install>`__ and
set the ``GOPATH``. You should also put ``GOPATH/bin`` on your ``PATH``.

Get Source Code
^^^^^^^^^^^^^^^

You should be able to install the latest with a simple

::

    go get github.com/teragrid/teragrid/cmd/teragrid

Run ``teragrid --help`` and ``teragrid version`` to ensure your
installation worked.

If the installation failed, a dependency may have been updated and become
incompatible with the latest teragrid master branch. We solve this
using the ``dep`` tool for dependency management.

First, install ``dep``:

::

    cd $GOPATH/src/github.com/teragrid/teragrid
    make get_tools

Now we can fetch the correct versions of each dependency by running:

::

    make get_vendor_deps
    make install

Note that even though ``go get`` originally failed, the repository was
still cloned to the correct location in the ``$GOPATH``.

The latest teragrid Core version is now installed.

Reinstall
---------

If you already have teragrid installed, and you make updates, simply

::

    cd $GOPATH/src/github.com/teragrid/teragrid
    make install

To upgrade, there are a few options:

-  set a new ``$GOPATH`` and run
   ``go get github.com/teragrid/teragrid/cmd/teragrid``. This
   makes a fresh copy of everything for the new version.
-  run ``go get -u github.com/teragrid/teragrid/cmd/teragrid``,
   where the ``-u`` fetches the latest updates for the repository and
   its dependencies
-  fetch and checkout the latest master branch in
   ``$GOPATH/src/github.com/teragrid/teragrid``, and then run
   ``make get_vendor_deps && make install`` as above.

Note the first two options should usually work, but may fail. If they
do, use ``dep``, as above:

::

    cd $GOPATH/src/github.com/teragrid/teragrid
    make get_vendor_deps
    make install

Since the third option just uses ``dep`` right away, it should always
work.

Troubleshooting
---------------

If ``go get`` failing bothers you, fetch the code using ``git``:

::

    mkdir -p $GOPATH/src/github.com/teragrid
    git clone https://github.com/teragrid/teragrid $GOPATH/src/github.com/teragrid/teragrid
    cd $GOPATH/src/github.com/teragrid/teragrid
    make get_tools
    make get_vendor_deps
    make install

Run
^^^

To start a one-node blockchain with a simple in-process application:

::

    teragrid init
    teragrid node --proxy_app=kvstore
