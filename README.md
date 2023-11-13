# Andy

Your generic programmer Andy, to automate some things, using Discord ✨✨✨

## Features

- Slash commands.
- Statically linked, just take the binary and run 🏃🏃🏃
- Compiled against six different targets, I'm running this on a Raspberry Pi 3 Model B.

## Planned

- Sandbox for running programs written in:
  - Go
  - Python
  - Fortran (yes, for my Numerical Analysis Lab 👴👴👴)

## Run

If you plan to host by yourself, first you will need a token. Follow these [instructions](ins),
you'll only need the first step, then make sure to give it the `Use Slash Commands` permission.

Now if you haven't already, then download the binary for your operating system and
processor's architecture from the [latest release](release) page.

After downloading, go to the downloaded folder, and unzip/decompress if you need to.
Then run the following:

```sh
./andy -token <TOKEN>
```

Replace `<TOKEN>` with the token you got from Discord.

## Development

You will need,

1. Go (version 1.20 or higher)
2. [Air 💨](air)

Then clone the repository,

```sh
git clone https://github.com/Fahim-Ferdous/andy
```

Enter the folder `andy`, and run,

```sh
go mod get
```

Now you can do all the changes you want to.

## Contribute

In addition to the steps in [Development](#development), you will also need the
following tools for commit messages,

1. [Commitlint](https://commitlint.js.org/)
2. [Husky](https://typicode.github.io/husky/)
3. [npm](https://nodejs.org/en/download), which comes bundled with NodeJS

After you install npm (NodeJS really), run the following in the folder where you have cloned Andy.
This will install Commitlint and Husky.

```sh
npm i
```

Send your PR's against the `main` branch and please note that the CI will reject
your PR if your commit messages do not pass the
[Conventional Commit](https://www.conventionalcommits.org/) standard of
[Semantic Release](ahttps://semantic-release.gitbook.io/)

To ensure your that your Pull Request will not get rejected, run

```sh
npx exec semantic-release
```

If you want contribute to the CI/CD pipeline, you'll additionally need the following,

1. [Act](https://github.com/nektos/act) (Note, you must choose the `Medium` image size)
2. [Docker 🐳](https://typicode.github.io/husky/) (Required for Act)

Now to verify that your PR will pass, run

```sh
act
```

You might get some warnings related to caching, but you can ignore those.
If you want to get rid of these warnings, [then this will help you](https://github.com/sp-ricard-valverde/github-act-cache-server).

[ins]: (https://discord.com/developers/docs/getting-started#step-1-creating-an-app)
[release]: (https://github.com/Fahim-Ferdous/andy/releases/latest)
[air]: (https://github.com/cosmtrek/air)
