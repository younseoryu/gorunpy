"""Entry point for the mathlib executable."""

# Import functions to register them with @export
import mathlib.functions  # noqa: F401

from gorunpy import main

if __name__ == "__main__":
    main()

