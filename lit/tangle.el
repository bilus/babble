;; Code generated from BOOK.org by make tangle. DO NOT EDIT.

;; Batch tangler for the org literate experiment. Supplies line-comment
;; syntax per language and turns the last prose paragraph before a
;; :comments org block into its comment.

(require 'org)
(require 'ob-tangle)
(require 'json)
(require 'ox)

;; Line-comment syntax, keyed by the language a block is written in. A
;; language absent here takes the hash, which comments shells, Python,
;; Ruby and make.
(defconst lit-comment-syntax
  '(("go" . "//") ("js" . "//")
    ("emacs-lisp" . ";;") ("elisp" . ";;")
    ("lua" . "--")))

(defvar lit-comment-table lit-comment-syntax)

;; A project extends the table through .lit/comments.json beside its
;; book, an object of language to prefix, whose entries win. The format
;; is JSON because both engines parse it with a reader neither of them
;; wrote, and a hand-written reader is where two implementations drift.
(defun lit-load-comment-table (book)
  (let ((file (expand-file-name
               ".lit/comments.json"
               (file-name-directory (expand-file-name book))))
        (table (copy-alist lit-comment-syntax)))
    (when (file-readable-p file)
      (dolist (pair (with-temp-buffer
                      (insert-file-contents file)
                      (json-parse-buffer :object-type 'alist)))
        (let ((lang (symbol-name (car pair)))
              (prefix (cdr pair)))
          (unless (stringp prefix)
            (error "%s: comment prefix for %s is not a string" file lang))
          (setf (alist-get lang table nil nil #'equal) prefix))))
    (setq lit-comment-table table)))

(defun lit-comment-prefix (lang)
  (or (cdr (assoc lang lit-comment-table)) "#"))

;; org takes comment syntax from a major mode, so a language whose mode
;; batch Emacs never loaded silently reuses the previous block's
;; syntax. Answering from the table instead makes comment syntax a
;; property of the language rather than of what Emacs happens to have
;; loaded.
(defun lit-lang-mode (lang)
  (let ((mode (intern (format "lit-comment-%s-mode" lang)))
        (prefix (lit-comment-prefix lang)))
    (unless (fboundp mode)
      (eval `(define-derived-mode ,mode prog-mode "lit"
               (setq-local comment-start ,prefix)
               (setq-local comment-end "")
               (setq-local comment-padding 1))))
    mode))

(advice-add 'org-src-get-lang-mode :override #'lit-lang-mode)

(setq comment-empty-lines t)

;; the post-tangle hooks modify and save every tangled file; without
;; this, batch Emacs strews foo~ backups through the tree
(when noninteractive
  (setq make-backup-files nil))

(setq org-babel-default-header-args
      (cons '(:mkdirp . "yes") org-babel-default-header-args))

;; The docstring for a :comments org block is the text after the last
;; "# doc" marker line when one is present, else the last paragraph.
;; Earlier text is narrative. A leading "Note: " or "Apropos: " sigil is
;; presentation, not part of the source comment.
(defun lit-docstring (text)
  (let* ((doc-end (let ((start 0) (end nil))
                    (while (string-match "^# doc\n" text start)
                      (setq end (match-end 0) start (match-end 0)))
                    end))
         (sliced (if doc-end (substring text doc-end) text))
         ;; #+name: and other org keyword/comment lines sit between the
         ;; docstring and the block; they are not comment material
         (clean (replace-regexp-in-string "^#\\+[^\n]*\n?" "" sliced))
         (chosen (if doc-end
                     clean
                   (car (last (split-string clean "\n\n+" t)))))
         (trimmed (string-trim (or chosen ""))))
    (replace-regexp-in-string "\\`\\(Note\\|Apropos\\): " "" trimmed)))

(setq org-babel-process-comment-text #'lit-docstring)

;; Which language wrote each target. org stamps the banner in a hook
;; that sees a filename and not a language, so the map is built while
;; the book is still open and read back by name. One file takes one
;; comment syntax, so two languages in a file is an error rather than
;; a guess.
(defvar lit-file-languages nil)

(defun lit-collect-file-languages ()
  (setq lit-file-languages nil)
  (dolist (group (org-babel-tangle-collect-blocks))
    (let ((file (expand-file-name (car group)))
          (langs (delete-dups (mapcar (lambda (b) (nth 0 b)) (cdr group)))))
      (when (cdr langs)
        (error "%s: blocks in %s; one file takes one comment syntax"
               file (mapconcat #'identity langs " and ")))
      (push (cons file (car langs)) lit-file-languages))))

(defun lit-file-comment-prefix ()
  (lit-comment-prefix
   (cdr (assoc (expand-file-name (buffer-file-name)) lit-file-languages))))

;; comment-region pads an empty comment line; the file wants the bare
;; marker.
(defun lit-trim-empty-comment-lines ()
  (let ((marker (lit-file-comment-prefix)))
    (goto-char (point-min))
    (while (re-search-forward (concat "^" (regexp-quote marker) " $") nil t)
      (replace-match marker))
    (save-buffer)))

(add-hook 'org-babel-post-tangle-hook #'lit-trim-empty-comment-lines)

;; Every tangled file opens with a generated-file banner in the comment
;; syntax of the language that wrote it, after a shebang when one is
;; present. The Go spelling is canonical ("Code generated ... DO NOT
;; EDIT.") so Go tooling recognizes it too.
(defconst lit-banner
  "Code generated from BOOK.org by make tangle. DO NOT EDIT.")

(defun lit-insert-banner ()
  (goto-char (point-min))
  (let ((banner (concat (lit-file-comment-prefix) " " lit-banner "\n")))
    (when (looking-at "#!.*\n")
      (goto-char (match-end 0)))
    (unless (looking-at-p (regexp-quote banner))
      (insert banner "\n"))
    (save-buffer)))

(add-hook 'org-babel-post-tangle-hook #'lit-insert-banner t)

;; A "#+begin: src-block-diff :old A :new B" dynamic block renders as
;; a unified diff of the two named blocks' bodies. The content is
;; derived: refresh one with C-c C-c on its begin line, and tangling
;; refreshes them all first, so a committed diff cannot disagree with
;; the blocks it describes.
(defun lit-named-block-body (name)
  (save-excursion
    (goto-char (or (org-babel-find-named-block name)
                   (error "no block named %s" name)))
    (nth 1 (org-babel-get-src-block-info t))))

(defun org-dblock-write:src-block-diff (params)
  ;; fetch both bodies before with-temp-file switches buffers
  (let ((old-body (lit-named-block-body
                   (format "%s" (plist-get params :old))))
        (new-body (lit-named-block-body
                   (format "%s" (plist-get params :new))))
        (old (make-temp-file "src-block-diff"))
        (new (make-temp-file "src-block-diff")))
    (unwind-protect
        (progn
          (with-temp-file old (insert old-body "\n"))
          (with-temp-file new (insert new-body "\n"))
          (let ((diff (with-temp-buffer
                        (call-process "diff" nil t nil "-u" old new)
                        (buffer-string))))
            (if (string= diff "")
                (insert "no differences")
              ;; the ---/+++ header lines only name the temp files
              (insert "#+begin_src diff\n"
                      (mapconcat #'identity
                                 (cddr (split-string diff "\n")) "\n")
                      "#+end_src"))))
      (delete-file old)
      (delete-file new))))

;; A "#+begin: plan-toc" dynamic block renders a jump list of every
;; plan-markup site in the buffer: the human twin of the census grep.
;; Labels are spelled without the sigils, so the census never matches
;; the index it feeds.
(defconst lit-plan-site-re
  (concat "^\\*+ \\(?:PLANNED\\|FILLING\\|HOLE\\|LIVE\\|DROPPED\\) "
          "\\|^\\*\\{15,\\} "
          "\\|HOLE("
          "\\|{\\+\\+\\|{--\\|{~~\\|{>>"
          "\\|^#\\+begin: src-block-diff"
          "\\|--planned"))

(defun lit-plan-strip (s)
  (let* ((s (string-trim s))
         (s (replace-regexp-in-string
             "\\[\\[[^]]*\\]\\[\\([^]]*\\)\\]\\]" "\\1" s))
         (s (replace-regexp-in-string "\\s-+:[a-zA-Z0-9_@:]+:$" "" s)))
    (if (> (length s) 60) (concat (substring s 0 57) "...") s)))

(defun lit-plan-site-label ()
  (let ((line (buffer-substring-no-properties
               (line-beginning-position) (line-end-position))))
    (cond
     ((string-match "^#\\+todo:" line) nil)
     ((string-match "^\\*\\{15,\\} END" line) nil)
     ((string-match "^\\*\\{15,\\} \\(.*\\)" line)
      (concat "badge: " (lit-plan-strip (match-string 1 line))))
     ((string-match "^\\*+ \\([A-Z]+ .*\\)" line)
      (concat "entry: " (lit-plan-strip (match-string 1 line))))
     ((string-match "^#\\+begin: src-block-diff" line) "generated diff")
     ((string-match "HOLE(\\([0-9]+\\)): \\([^\"]*\\)" line)
      (format "hole %s: %s" (match-string 1 line)
              (lit-plan-strip (match-string 2 line))))
     ((string-match "{\\+\\+\\(.*\\)" line)
      (concat "insertion: " (lit-plan-strip (match-string 1 line))))
     ((string-match "{--\\(.*\\)" line)
      (concat "deletion: " (lit-plan-strip (match-string 1 line))))
     ((string-match "{~~\\(.*\\)" line)
      (concat "replacement: " (lit-plan-strip (match-string 1 line))))
     ((string-match "{>>\\(.*\\)" line)
      (concat "note: " (lit-plan-strip (match-string 1 line))))
     ((string-match "#\\+name: \\(.*\\)--planned" line)
      (concat "twin of " (match-string 1 line)))
     (t nil))))

(defun org-dblock-write:plan-toc (_params)
  (let ((file (file-name-nondirectory (buffer-file-name)))
        (self (line-number-at-pos))
        (hits ()))
    (save-excursion
      (goto-char (point-min))
      (while (re-search-forward lit-plan-site-re nil t)
        (beginning-of-line)
        (let ((label (lit-plan-site-label)))
          (when label
            (push (cons (line-number-at-pos) label) hits)))
        (forward-line 1)))
    (setq hits (nreverse hits))
    ;; sites below this block shift down by the lines inserted here
    (let ((shift (length hits)))
      (dolist (h hits)
        (let ((n (if (> (car h) self) (+ (car h) shift) (car h))))
          (insert (format "- [[file:%s::%d][%d]] %s\n"
                          file n n (cdr h))))))))

(defun lit-check-includes ()
  (goto-char (point-min))
  (while (re-search-forward "^[ \t]*#\\+INCLUDE:\\(.*\\)$" nil t)
    (let ((rest (string-trim (match-string 1))))
      (unless (string-match "\\`\"[^\"]+\"\\(\\s-+example\\)?\\'" rest)
        (error "%s:%d: only #+INCLUDE: \"path\" and #+INCLUDE: \"path\" \
example are understood"
               (buffer-file-name) (line-number-at-pos))))))

(defun lit-neutralize-example-includes ()
  "Rename wrapped includes so the expander passes over them.
The buffer is thrown away after tangling, and the line stays a
keyword line, which the docstring rule drops either way."
  (goto-char (point-min))
  (while (re-search-forward
          "^\\([ \t]*\\)#\\+INCLUDE:\\([ \t]*\"[^\"]+\"[ \t]+example[ \t]*\\)$"
          nil t)
    (replace-match "\\1#+example_include:\\2")))

(defun lit-tangle (file)
  (with-current-buffer (find-file-noselect file)
    (org-update-all-dblocks)
    (when (buffer-modified-p)
      (save-buffer))
    (lit-check-includes)
    (lit-load-comment-table file)
    ;; Expanding rewrites the buffer, so nothing may save it after this
    ;; point: org-babel-pre-tangle-hook saves by default, and the master
    ;; on disk would become the merged book.
    (lit-neutralize-example-includes)
    (org-export-expand-include-keyword)
    (lit-collect-file-languages)
    (let ((org-babel-pre-tangle-hook nil))
      (org-babel-tangle))
    (set-buffer-modified-p nil)))
