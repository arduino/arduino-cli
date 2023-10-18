# Source:
# https://github.com/arduino/tooling-project-assets/blob/main/workflow-templates/assets/deploy-mkdocs-versioned/siteversion/siteversion.py

# Copyright 2020 ARDUINO SA (http://www.arduino.cc/)

# This software is released under the GNU General Public License version 3
# The terms of this license can be found at:
# https://www.gnu.org/licenses/gpl-3.0.en.html

# You can be released from the requirements of the above licenses by purchasing
# a commercial license. Buying such a license is mandatory if you want to
# modify or otherwise use the software for commercial activities involving the
# Arduino software without disclosing the source code of your own applications.
# To purchase a commercial license, send an email to license@arduino.cc.
import os
import re
import json

from git import Repo

# In order to provide support for multiple project releases, Documentation is versioned so that visitors can select
# which version of the documentation website should be displayed. Unfortunately this feature isn't provided by GitHub
# pages or MkDocs, so we had to implement it on top of the generation process.
#
# - A special version of the documentation called `dev` is provided to reflect the status of the project on the
#   default branch - this includes unreleased features and bugfixes.
# - Docs are versioned after the minor version of a release. For example, release version `0.99.1` and
#   `0.99.2` will be both covered by documentation version `0.99`.
#
# The CI is responsible for guessing which version of the project we're building docs for, so that generated content
# will be stored in the appropriate section of the documentation website. Because this guessing might be fairly complex,
# the logic is implemented in this Python script. The script will determine the version of the project that was
# modified in the current commit (either `dev` or an official, numbered release) and whether the redirect to the latest
# version that happens on the landing page should be updated or not.


DEV_BRANCHES = ["master"]  # Name of the branch used for the "dev" website source content


def get_docs_version(ref_name, release_branches):
    if ref_name in DEV_BRANCHES:
        return {"version": "dev", "alias": ""}

    if ref_name in release_branches:
        # if version is latest, add an alias
        alias = "latest" if ref_name == release_branches[0] else ""
        # strip `.x` suffix from the branch name to get the version: 0.3.x -> 0.3
        return {"version": ref_name[:-2], "alias": alias}

    return {"version": None, "alias": None}


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


def main():
    # Detect repo root folder
    here = os.path.dirname(os.path.realpath(__file__))
    repo_dir = os.path.join(here, "..", "..")

    # Get current repo
    repo = Repo(repo_dir)

    # Get the list of release branch names
    rel_br_names = get_rel_branch_names(repo.refs)

    # Deduce docs version from current branch.
    versioning_data = get_docs_version(repo.active_branch.name, rel_br_names)

    # Return the data as JSON on stdout
    print(json.dumps(versioning_data))


# Usage:
#     To run the script (must be run from within the repo tree):
#         $python siteversion.py
#
if __name__ == "__main__":
    main()
