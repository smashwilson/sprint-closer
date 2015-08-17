# Sprint Closer

This is a command-line tool that automates the process of closing out our sprints in Trello each week.

 * Creates a new board in the DevEx organization with a name containing the date of the previous Friday.
 * Add each member of the organization to the new board and clean out the pre-existing lists.
 * Move the "done" list from the current sprint to the archive board.
 * Create a new, empty "done" list on the current sprint board in the same position that the old one was.

## Installation and Configuration

To install:

 1. Download the binary from the [latest release on GitHub](https://github.com/smashwilson/sprint-closer/releases).
 2. Use `chmod +x ./sprint-closer` to make it executable.
 3. *(Optional)* Move the sprint-closer binary somewhere on your `${PATH}`, so that you can run it without the `./` prefix.

Now, you'll need to perform some one-time configuration by creating a file in your home directory with some connection information. First, you'll need to generate a *key* and a *token* from your Trello account.

 1. To generate your **key**, make sure that you're logged in to Trello, then visit [https://trello.com/1/appKey/generate](https://trello.com/1/appKey/generate).
 2. Now, use that key to generate your **token** by visiting: `https://trello.com/1/authorize?key=${KEY}&name=Closer&expiration=never&scope=read,write&response_type=token`.
 3. Find your organization's Trello name by visiting the organization's page and checking the URL:

![org-name](https://cloud.githubusercontent.com/assets/17565/9306213/d5d1681e-44c4-11e5-87a0-72ba59a6b11f.jpeg)

Put these three pieces of information in a JSON file called `~/.trello.json`:

```json
{
  "key": "...",
  "token": "...",
  "organization": "automationtesting2"
}
```

## Usage

To close the sprint each week, run:

```bash
sprint-closer

# Or ./sprint-closer if it's not on your ${PATH}
```

:sparkles:

If something goes wrong or you want more details about what it's doing, you can crank up the logging level with:

```bash
sprint-closer --log debug
```

Finally, this is probably not relevant unless you're developing sprint-closer itself, but you can use a different path for the Trello configuration:

```bash
sprint-closer --profile ~/somewhereelse/testaccount.json
```
