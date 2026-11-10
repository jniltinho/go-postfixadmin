Statuses concern the **revised plan**, not implementation. Paths are relative to the change directory.

1. **fixed** — XP/98 now use responsive sizing (`design.md:34`).
2. **fixed** — Scoped pixel typography/control migration added (`tasks.md:44`).
3. **fixed** — User portal layout explicitly included (`tasks.md:43`).
4. **fixed** — Pixel equality limited to the theme-only refactor; missing baselines scheduled (`specs/ui-theming/spec.md:11`, `tasks.md:8`).
5. **fixed** — `admanager` added to the catalog (`specs/ui-theming/spec.md:15`).
6. **fixed** — Anti-flash script now applies `color-scheme` too (`design.md:40`).
7. **fixed** — Preference, invalid-ID, precedence and API-failure tests added (`tasks.md:36`).
8. **fixed** — Generated Vite configs removed, ignored and prevented from reappearing (`tasks.md:21`).
9. **fixed** — HTTP regression tests required before upgrading Echo (`tasks.md:9`).
10. **fixed** — Swagger regeneration added after `/version` changes (`tasks.md:35`).
11. **fixed** — Manual SHA-only tags and stable-only `latest` specified (`design.md:59,83`).
12. **fixed** — Public visibility setup and anonymous-pull verification added (`tasks.md:53`).
13. **not fixed** — README claim removed, but multi-arch package descriptions still rely only on image labels (`design.md:59`); require the manifest-index description annotation. [GHCR documentation](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry#adding-a-description-to-multi-arch-images).
**New findings**
14. **`tasks.md:20`** — Copying ad-manager’s `vite` override conflicts with this repo’s direct `vite: "^8.0.12"` dependency (`frontend/package.json:26`), causing `EOVERRIDE`; explicitly remove or re-alias that dependency. [npm rules](https://docs.npmjs.com/cli/v11/configuring-npm/package-json/#overrides).
15. **`specs/container-image-publishing/spec.md:4` versus `design.md:59`** — Publishing promises every `v*` tag, but semver-only tagging produces no tag for `vnext`; restrict the promise to valid semver or add a fallback. [Metadata-action requirements](https://github.com/docker/metadata-action#typesemver).
No files edited. Builds, browser tests and publishing were not run; runtime behavior remains unverified.
99.070
