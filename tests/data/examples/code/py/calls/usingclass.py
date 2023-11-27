import os
import sys
import fnmatch
from collections import defaultdict

class CodeParser:
    def __init__(self):
        self.repo_analysis = defaultdict(list)
        self.direct_dependencies = set()

    @staticmethod
    def find_direct_dependencies(self, dirpath, include_extensions, exclude_dirs):
        root_packages = self.find_top_level_modules(dirpath)
        self.find_modules_recursive(dirpath, include_extensions, exclude_dirs)
        self.find_unique_modules(root_packages)

    def find_top_level_modules(self, dirpath):
        root_packages = {}
        # Implement code to find top-level modules and populate root_packages dictionary
        return root_packages

    def find_unique_modules(self, root_packages):
        unique_mod_names = set()
        for file_analysis_list in self.repo_analysis.values():
            for module in file_analysis_list:
                # Implement code to check for relative and local imports and extract top-level packages
                is_relative_import = False
                is_import_local = False

                for pkg in root_packages:
                    is_relative_import = is_relative_import or module.name.startswith(".")
                    is_import_local = is_import_local or module.name.startswith(pkg + ".") or module.name == pkg

                if not is_relative_import and not is_import_local:
                    top_level_pkg = module.name.split(".")[0]
                    self.direct_dependencies.add(top_level_pkg)
                    unique_mod_names.add(top_level_pkg)

        return self.direct_dependencies

    def find_modules_recursive(self, dirpath, include_extensions, exclude_dirs):
        for root, _, files in os.walk(dirpath):
            if self.should_exclude_dir(root, exclude_dirs):
                continue

            for file in files:
                if self.should_include_file(file, include_extensions):
                    file_path = os.path.join(root, file)
                    self.find_modules_in_file(file_path, dirpath)

    def should_exclude_dir(self, dir_path, exclude_dirs):
        for exclude_dir in exclude_dirs:
            if dir_path == exclude_dir:
                return True
        return False

    def should_include_file(self, file_name, include_extensions):
        file_extension = os.path.splitext(file_name)[1]
        return file_extension in include_extensions

    def find_modules_in_file(self, file_path, root_dir):
        # Implement code to parse the code file and extract modules
        pass


def find_direct_dependencies(dirpath, include_extensions, exclude_dirs):
    repo_analysis = defaultdict(list)
    direct_dependencies = set()
    root_packages = find_top_level_modules(dirpath)
    find_modules_recursive(dirpath, include_extensions, exclude_dirs)
    direct_dependencies = find_unique_modules(root_packages, repo_analysis)
    return direct_dependencies

def find_top_level_modules(dirpath):
    root_packages = {}
    # Implement code to find top-level modules and populate root_packages dictionary
    return root_packages

def find_unique_modules(root_packages, repo_analysis):
    unique_mod_names = set()
    for file_analysis_list in repo_analysis.values():
        for module in file_analysis_list:
            # Implement code to check for relative and local imports and extract top-level packages
            is_relative_import = False
            is_import_local = False

            for pkg in root_packages:
                is_relative_import = is_relative_import or module.name.startswith(".")
                is_import_local = is_import_local or module.name.startswith(pkg + ".") or module.name == pkg

            if not is_relative_import and not is_import_local:
                top_level_pkg = module.name.split(".")[0]
                direct_dependencies.add(top_level_pkg)
                unique_mod_names.add(top_level_pkg)

    return direct_dependencies

def find_modules_recursive(dirpath, include_extensions, exclude_dirs):
    for root, _, files in os.walk(dirpath):
        if should_exclude_dir(root, exclude_dirs):
            continue

        for file in files:
            if should_include_file(file, include_extensions):
                file_path = os.path.join(root, file)
                find_modules_in_file(file_path, dirpath)

def should_exclude_dir(dir_path, exclude_dirs):
    for exclude_dir in exclude_dirs:
        if dir_path == exclude_dir:
            return True
    return False

def should_include_file(file_name, include_extensions):
    file_extension = os.path.splitext(file_name)[1]
    return file_extension in include_extensions

def find_modules_in_file(file_path, root_dir):
    # Implement code to parse the code file and extract modules
    pass

# Example usage:
dir_path = "/path/to/your/code/directory"
include_extensions = [".py"]
exclude_dirs = ["/path/to/exclude/dir1", "/path/to/exclude/dir2"]
direct_dependencies = find_direct_dependencies(dir_path, include_extensions, exclude_dirs)
print("Direct Dependencies:", direct_dependencies)

# Example usage:
code_parser = CodeParser()
dir_path = "/path/to/your/code/directory"
include_extensions = [".py"]
exclude_dirs = ["/path/to/exclude/dir1", "/path/to/exclude/dir2"]
code_parser.find_direct_dependencies(dir_path, include_extensions, exclude_dirs)
direct_dependencies = code_parser.direct_dependencies
print("Direct Dependencies:", direct_dependencies)
