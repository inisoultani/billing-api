# Billing API Collection

This repository is a **Bruno** collection. Bruno is an offline-first, Git-friendly API client that stores everything as plain text files directly in this folder.

## Prerequisites

1. **The App:** [Download Bruno Desktop](https://www.usebruno.com/downloads) (or use the VS Code extension).
2. **No Import Needed:** Do **not** use the "Import" button. Simply **Open** the folder.

---

## Quick Start

### 1. Open the Collection

1. Launch Bruno.
2. Click **"Open Collection"** on the home screen.
3. Select the `billing-api-collection` folder.

- _Bruno will automatically detect all requests and environments within the folder._

### 2. Activate the Environment

To use the shared variables (protocol, host, port):

1. Locate the dropdown in the **top-right corner** (defaults to "No Environment").
2. Select **`local`**.
3. All requests will now resolve to: `{{protocol}}://{{host}}:{{port}}/loan/:loanID`

### 3. Path Parameters

For URLs containing `:loanID`, Bruno will automatically provide a field in the **Vars** tab for you to enter the specific ID you want to test.

---
