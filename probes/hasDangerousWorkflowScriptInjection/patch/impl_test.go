package hasDangerousWorkflowScriptInjection

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_Run(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input      string
		expected string
		err      error
	}{
		// Extracted from real Angular fix: https://github.com/angular/angular/pull/51026/files
		{	
			name: "var: Comment Body",
			input:
			`
			name: Run benchmark comparison

			on:
				issue_comment:
					types: [created]
			
			jobs:
			  banchmark-compare:
			  	steps:
					- name: Preparing benchmark for GitHub action
						id: info
						run: yarn benchmarks prepare-for-github-action "${{github.event.comment.body}}"
			`,
			expected:
			`
			name: Run benchmark comparison

			on:
				issue_comment:
					types: [created]
			
			jobs:
			  banchmark-compare:
			  	steps:
					- name: Preparing benchmark for GitHub action
						id: info
						env:
							COMMENT_BODY: ${{ github.event.comment.body }}
					    run: yarn benchmarks prepare-for-github-action "$COMMENT_BODY"
			`,
		},
		// Inspired on a real fix: https://github.com/googleapis/google-cloud-go/pull/9011/files
		{	
			name: "var: Pull Request Head Ref. Also uses untrusted input inside IF",
			input:
			`
			name: apidiff
			on:
				pull_request:
			
			permissions:
				contents: read
				pull-requests: write
			
			jobs:
				apidiff:
					needs: scan_changes
					runs-on: ubuntu-latest
					if: ${{ !needs.scan_changes.outputs.skip && !contains(github.event.pull_request.labels.*.name, 'breaking change allowed') }}
					continue-on-error: true
					strategy:
						matrix: ${{ fromJson(needs.scan_changes.outputs.changed_dirs) }}
					steps:
						- name: Compare regenerated code to baseline
						  run: |
							cd ${{ matrix.changed }} && apidiff -m -incompatible ${{ steps.baseline.outputs.pkg }} . > diff.txt
							if [[ ${{ github.event.pull_request.head.ref }} == owl-bot-copy ]]; then
								sed -i '/: added/d' ./diff.txt
							fi
							cat diff.txt && ! [ -s diff.txt ]
			`,
			expected:
			`
			name: apidiff
			on:
				pull_request:
			
			permissions:
				contents: read
				pull-requests: write
			
			jobs:
				apidiff:
					needs: scan_changes
					runs-on: ubuntu-latest
					if: ${{ !needs.scan_changes.outputs.skip && !contains(github.event.pull_request.labels.*.name, 'breaking change allowed') }}
					continue-on-error: true
					strategy:
						matrix: ${{ fromJson(needs.scan_changes.outputs.changed_dirs) }}
					steps:
						- name: Compare regenerated code to baseline
						  env:
							PULL_REQUEST_HEAD_REF: ${{ github.event.pull_request.head.ref }}
						  run: |
							cd ${{ matrix.changed }} && apidiff -m -incompatible ${{ steps.baseline.outputs.pkg }} . > diff.txt
							if [[ "$PR_HEAD_REF" == owl-bot-copy ]]; then
								sed -i '/: added/d' ./diff.txt
							fi
							cat diff.txt && ! [ -s diff.txt ]
			`,
		},
		// Inspired from a real lit/lit fix: https://github.com/lit/lit/pull/3669/files
		{	
			name: "var: Pull Request Body",
			input:
			`
			name: Generate Release Image

			on:
				pull_request:
					paths:
					- '**/CHANGELOG.md'
			
			jobs:
				release-image:
					steps:
						- name: Create release image
						  run: |
						    echo "${{ github.event.pull_request.body }}" > release.md
			`,
			expected:
			`
			name: Generate Release Image

			on:
				pull_request:
					paths:
					- '**/CHANGELOG.md'
			
			jobs:
				release-image:
					steps:
						- name: Create release image
						  env:
						  	PULL_REQUEST_BODY: ${{ github.event.pull_request.body }}
						  run: |
						    echo "$PULL_REQUEST_BODY" > release.md
			`,
		},
		{	
			name: "Two incidences in different jobs. Vars: issue title and issue body",
			input:
			`
			on:
				issue:
			
			jobs:
				fascinating-job:
					steps:
						- name: it runs like magic
						  run: |
						    echo "${{ github.event.issue.title }}"

				incredible-other-job:
					steps:
						- name: absolutely outstanding, safe as nothing else
						  run: |
						    echo "${{ github.event.issue.body }}" 
			`,
			expected:
			`
			on:
				issue:
			
			jobs:
				fascinating-job:
					steps:
						- name: it runs like magic
						  env:
							ISSUE_TITLE: "${{ github.event.issue.title }}"
						  run: |
						    echo "$ISSUE_TITLE"

				incredible-other-job:
					steps:
						- name: absolutely outstanding, safe as nothing else
						  env:
						    ISSUE_BODY: ${{ github.event.issue.body }}
						  run: |
							echo "$ISSUE_BODY" 
			`,
		},
		{	
			name: "Two incidences in same job. Vars: discussion title and discussion body",
			input:
			`
			on:
				discussion:
					types: [created]
			
			jobs:
				really-complete-job:
					steps:
						- name: it's only the beginning
						  run: |
						    echo "${{ github.event.discussion.title }}"
						- name: ok, now we're talking
						  run: |
						    echo "${{ github.event.discussion.body }}"
			`,
			expected:
			`
			on:
				discussion:
					types: [created]
			
			jobs:
				really-complete-job:
					steps:
						- name: it's only the beginning
						  env:
						    DISCUSSION_TITLE: ${{ github.event.discussion.title }}
						  run: |
						    echo "$DISCUSSION_TITLE"
						- name: ok, now we're talking
						  env:
						    DISCUSSION_BODY: ${{ github.event.discussion.body }}
						  run: |
						    echo "$DISCUSSION_BODY"
			`,
		},
		{	
			name: "Two incidences in same step. Vars: issue comment and fork name",
			input:
			`
			on:
				fork
				issue_comment:
					types: [created, edited]
			jobs:
				solution-to-all-repo-problems:
					steps:
						- name: where things are done
						  run: |
						    echo "${{ github.event.issue_comment.comment }}"
							mkdir "${{ github.event.fork.forkee.name }}"
			`,
			expected:
			`
			on:
				fork
				issue_comment:
					types: [created, edited]
			jobs:
				solution-to-all-repo-problems:
					steps:
						- name: where things are done
						  env:
						  	ISSUE_COMMENT_COMMENT: ${{ github.event.issue_comment.comment }}
							FORK_FORKEE_NAME: ${{ github.event.fork.forkee.name }}
						  run: |
						    echo "$ISSUE_COMMENT_COMMENT"
							mkdir "$FORK_FORKEE_NAME"
			`,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			output := Run(tt.input)
			if diff := cmp.Diff(tt.expected, output); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}