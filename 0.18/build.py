# This file is part of arduino-cli.

# Copyright 2020 ARDUINO SA (http://www.arduino.cc/)

# This software is released under the GNU General Public License version 3,
# which covers the main part of arduino-cli.
# The terms of this license can be found at:
# https://www.gnu.org/licenses/gpl-3.0.en.html

# You can be released from the requirements of the above licenses by purchasing
# a commercial license. Buying such a license is mandatory if you want to
# modify or otherwise use the software for commercial activities involving the
# Arduino software without disclosing the source code of your own applications.
# To purchase a commercial license, send an email to license@arduino.cc.
import os
import sys
import re
import unittest
import subprocess

import click
from git import Repo


DEV_BRANCHES = ["master"]


class TestScript(unittest.TestCase):
    def test_get_docs_version(self):
        ver, alias = get_docs_version("master", [])
        self.assertEqual(ver, "dev")
        self.assertEqual(alias, "")

        release_names = ["1.4.x", "0.13.x"]
        ver, alias = get_docs_version("0.13.x", release_names)
        self.assertEqual(ver, "0.13")
        self.assertEqual(alias, "")
        ver, alias = get_docs_version("1.4.x", release_names)
        self.assertEqual(ver, "1.4")
        self.assertEqual(alias, "latest")

        ver, alias = get_docs_version("0.1.x", [])
        self.assertIsNone(ver)
        self.assertIsNone(alias)


def get_docs_version(ref_name, release_branches):
    if ref_name in DEV_BRANCHES:
        return "dev", ""

    if ref_name in release_branches:
        # if version is latest, add an alias
        alias = "latest" if ref_name == release_branches[0] else ""
        # strip `.x` suffix from the branch name to get the version: 0.3.x -> 0.3
        return ref_name[:-2], alias

    return None, None


def get_rel_branch_names(blist):
    """Get the names of the release branches, sorted from newest to older.

    Only process remote refs so we're sure to get all of them and clean up the
    name so that we have a list of strings like 0.6.x, 0.7.x, ...
    """
    pattern = re.compile(r"origin/(\d+\.\d+\.x)")
    names = []
    for b in blist:
        res = pattern.search(b.name)
        if res is not None:
            names.append(res.group(1))

    # Since sorting is stable, first sort by major...
    names = sorted(names, key=lambda x: int(x.split(".")[0]), reverse=True)
    # ...then by minor
    return sorted(names, key=lambda x: int(x.split(".")[1]), reverse=True)


@click.command()
@click.option("--test", is_flag=True)
@click.option("--dry", is_flag=True)
@click.option("--remote", default="origin", help="The git remote where to push.")
def main(test, dry, remote):
    # Run tests if requested
    if test:
        unittest.main(argv=[""], exit=False)
        sys.exit(0)

    # Detect repo root folder
    here = os.path.dirname(os.path.realpath(__file__))
    repo_dir = os.path.join(here, "..")

    # Get current repo
    repo = Repo(repo_dir)

    # Get the list of release branch names
    rel_br_names = get_rel_branch_names(repo.refs)

    # Deduce docs version from current branch. Use the 'latest' alias if
    # version is the most recent
    docs_version, alias = get_docs_version(repo.active_branch.name, rel_br_names)
    if docs_version is None:
        print(f"Can't get version from current branch '{repo.active_branch}', skip docs generation")
        return 0

    # Taskfile args aren't regular args so we put everything in one string
    cmd = (f"task docs:publish DOCS_REMOTE={remote} DOCS_VERSION={docs_version} DOCS_ALIAS={alias}",)

    if dry:
        print(cmd)
        return 0

    subprocess.run(cmd, shell=True, check=True, cwd=repo_dir)


# Usage:
#
#     To run the tests:
#         $python build.py test
#
#     To run the script (must be run from within the repo tree):
#         $python build.py
#
if __name__ == "__main__":
    sys.exit(main())
