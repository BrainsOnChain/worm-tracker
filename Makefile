.PHONY: confirm
_WARN := "\033[33m[%s]\033[0m %s\n"  # Yellow text for "printf"
_TITLE := "\033[32m[%s]\033[0m %s\n" # Green text for "printf"
_ERROR := "\033[31m[%s]\033[0m %s\n" # Red text for "printf"

CURRENT_BRANCH = $(shell git branch --show-current) 
COMMIT = $(shell git rev-parse --short=12 HEAD)



# ------------------------------------------------------------------------------
# Development

deploy:
	@echo "Deploying to fly.io dev"
	fly deploy -c fly.toml

logs:
	@echo "Showing logs for dev"
	fly logs -a worm-tracker

# print-releases lists the last 5 releases for the dev deployment
print-releases:
	fly releases -a worm-tracker --image --json | jq 'limit(5; .[]) | {Version, Description, ImageRef, CreatedAt, UserEmail: .User.Email}'

# rollback rolls back the dev deployment to the specified IMAGE
rollback:
	@echo "Rolling back dev to ${IMAGE}"
	fly deploy -a worm-tracker --image ${IMAGE}


# ------------------------------------------------------------------------------
# Helpers

# Enforce the current branch is main
main-required:
	make branch-check CHECK_BRANCH="main"

# Check that the current branch is the provided CHECK_BRANCH
branch-check:
	@if [ "$$(git branch --show-current)" != "$(CHECK_BRANCH)" ]; then	\
		echo "$(tput setaf 3)WARNING: Current git branch is not $(CHECK_BRANCH): $$(git branch --show-current)"; \
		exit 1; \
	fi

# The CI environment variable can be set to a non-empty string,
# it'll bypass this command that will "return true", as a "yes" answer.
confirm:
	@if [[ -z "$(CI)" ]]; then \
		REPLY="" ; \
		read -p "⚠ Are you sure? [y/n] > " -r ; \
		if [[ ! $$REPLY =~ ^[Yy]$$ ]]; then \
			printf $(_ERROR) "KO" "Stopping" ; \
			exit 1 ; \
		else \
			printf $(_TITLE) "OK" "Continuing" ; \
			exit 0; \
		fi \
	fi

	