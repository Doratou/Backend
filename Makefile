.PHONY: up down run stop migrate-up migrate-down db-reset

# Pick the script folder that matches the OS.
#   Windows     -> scripts/windows/<target>.ps1
#   Linux/macOS -> scripts/linux/<target>.sh
ifeq ($(OS),Windows_NT)
  SCRIPT_DIR := scripts/windows
  EXT := ps1
  RUNNER := powershell -NoProfile -ExecutionPolicy Bypass -File
else
  SCRIPT_DIR := scripts/linux
  EXT := sh
  RUNNER := bash
endif

# Each target runs the script with the same name.
# Examples:  make up | make migrate-down STEPS=2 | make db-reset YES=1
up down run stop migrate-up migrate-down db-reset:
	@$(RUNNER) $(SCRIPT_DIR)/$@.$(EXT)
