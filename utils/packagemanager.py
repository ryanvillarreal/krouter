import apt
import apt.progress
import logging


class PackageExistsError(Exception):
	''' Custom error msg.'''
	def __init__(self, message):
		self.message = message

class PackageManager:
	def __init__(self):
		self.cache = apt.Cache()
		self.new_packages = []


	def update_cache(self):
		self.cache.update()
		self.cache.open(None)


	def install_package(self, package):
		try:
			package = self.cache[package]
		except KeyError as e:
			logging.warning(f"Package not found in cache: {package}")
		try:
			if not package.is_installed:
				package.mark_install()
				self.new_packages.append(package)
				self.cache.commit()
			else:
				raise PackageExistsError(f"Package already installed: {package}")
		except Exception as e:
			print(f"{e}")


	def uninstall_package(self, package):
		package = self.cache[package]
		package.mark_delete()
		self.cache.commit()


	def upgrade_packages(self):
		self.cache.upgrade()
		self.cache.commit()