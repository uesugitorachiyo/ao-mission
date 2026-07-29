package mission

import (
	"bytes"
	"compress/zlib"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	soakCanaryGitMaxIndexBytes   = 64 << 20
	soakCanaryGitMaxEntries      = 200_000
	soakCanaryGitMaxObjectBytes  = 64 << 20
	soakCanaryGitMaxPackIndex    = 256 << 20
	soakCanaryGitMaxPacks        = 128
	soakCanaryGitMaxDeltaDepth   = 64
	soakCanaryGitMaxReferenceLen = 4 << 10
)

type soakCanaryGitLayout struct {
	root      string
	gitDir    string
	commonDir string
}

type soakCanaryGitOID [sha1.Size]byte

type soakCanaryGitIndexEntry struct {
	path string
	mode uint32
	oid  soakCanaryGitOID
}

type soakCanaryGitTreeEntry struct {
	mode uint32
	oid  soakCanaryGitOID
}

type soakCanaryGitObjectStore struct {
	commonDir string
}

type soakCanaryGitObject struct {
	kind string
	data []byte
}

func (InProcessSoakCanaryGitVerifier) Verify(repositoryRoot, expectedRevision string) error {
	layout, err := resolveSoakCanaryGitLayout(repositoryRoot)
	if err != nil {
		return err
	}
	expected, err := parseSoakCanaryGitOID(expectedRevision)
	if err != nil {
		return errors.New("soak canary approved Git revision is invalid")
	}
	head, err := resolveSoakCanaryGitHEAD(layout)
	if err != nil {
		return err
	}
	if head != expected {
		return fmt.Errorf("soak canary Git HEAD=%s want=%s", head, expected)
	}
	index, err := loadSoakCanaryGitIndex(layout)
	if err != nil {
		return err
	}
	store := soakCanaryGitObjectStore{commonDir: layout.commonDir}
	tree, err := loadSoakCanaryGitHEADTree(store, head)
	if err != nil {
		return err
	}
	if err := compareSoakCanaryGitIndexToHEAD(index, tree); err != nil {
		return err
	}
	fileMode, err := soakCanaryGitCoreFileMode(layout)
	if err != nil {
		return err
	}
	if err := verifySoakCanaryGitWorktree(layout.root, index, fileMode); err != nil {
		return err
	}
	return rejectSoakCanaryGitUntracked(layout.root, index)
}

func resolveSoakCanaryGitLayout(repositoryRoot string) (soakCanaryGitLayout, error) {
	root, err := filepath.Abs(repositoryRoot)
	if err != nil {
		return soakCanaryGitLayout{}, err
	}
	info, err := os.Lstat(root)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return soakCanaryGitLayout{}, errors.New("soak canary Git root is not a regular directory")
	}
	dotGit := filepath.Join(root, ".git")
	dotInfo, err := os.Lstat(dotGit)
	if err != nil {
		return soakCanaryGitLayout{}, errors.New("soak canary .git metadata is missing")
	}
	var gitDir string
	switch {
	case dotInfo.IsDir() && dotInfo.Mode()&os.ModeSymlink == 0:
		gitDir = dotGit
	case dotInfo.Mode().IsRegular() && dotInfo.Mode()&os.ModeSymlink == 0:
		body, err := readBoundedRegularFile(dotGit, soakCanaryGitMaxReferenceLen)
		if err != nil {
			return soakCanaryGitLayout{}, err
		}
		line := strings.TrimSpace(string(body))
		if !strings.HasPrefix(line, "gitdir: ") || strings.Contains(line, "\n") {
			return soakCanaryGitLayout{}, errors.New("soak canary Git file is malformed")
		}
		gitDir = strings.TrimSpace(strings.TrimPrefix(line, "gitdir: "))
		if !filepath.IsAbs(gitDir) {
			gitDir = filepath.Join(root, gitDir)
		}
	default:
		return soakCanaryGitLayout{}, errors.New("soak canary .git metadata is unsupported")
	}
	gitDir, err = cleanSoakCanaryGitDirectory(gitDir)
	if err != nil {
		return soakCanaryGitLayout{}, err
	}
	commonDir := gitDir
	commonPath := filepath.Join(gitDir, "commondir")
	if info, statErr := os.Lstat(commonPath); statErr == nil {
		if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return soakCanaryGitLayout{}, errors.New("soak canary Git commondir is unsafe")
		}
		body, err := readBoundedRegularFile(commonPath, soakCanaryGitMaxReferenceLen)
		if err != nil {
			return soakCanaryGitLayout{}, err
		}
		value := strings.TrimSpace(string(body))
		if value == "" || strings.Contains(value, "\n") {
			return soakCanaryGitLayout{}, errors.New("soak canary Git commondir is malformed")
		}
		if !filepath.IsAbs(value) {
			value = filepath.Join(gitDir, value)
		}
		commonDir, err = cleanSoakCanaryGitDirectory(value)
		if err != nil {
			return soakCanaryGitLayout{}, err
		}
	} else if !os.IsNotExist(statErr) {
		return soakCanaryGitLayout{}, statErr
	}
	return soakCanaryGitLayout{root: root, gitDir: gitDir, commonDir: commonDir}, nil
}

func cleanSoakCanaryGitDirectory(path string) (string, error) {
	absolute, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return "", err
	}
	info, err := os.Lstat(absolute)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "", errors.New("soak canary Git metadata directory is unsafe")
	}
	return absolute, nil
}

func resolveSoakCanaryGitHEAD(layout soakCanaryGitLayout) (soakCanaryGitOID, error) {
	body, err := readBoundedRegularFile(
		filepath.Join(layout.gitDir, "HEAD"),
		soakCanaryGitMaxReferenceLen,
	)
	if err != nil {
		return soakCanaryGitOID{}, fmt.Errorf("read soak canary Git HEAD: %w", err)
	}
	value := strings.TrimSpace(string(body))
	if strings.HasPrefix(value, "ref: ") {
		return resolveSoakCanaryGitReference(
			layout,
			strings.TrimSpace(strings.TrimPrefix(value, "ref: ")),
			0,
		)
	}
	oid, err := parseSoakCanaryGitOID(value)
	if err != nil {
		return soakCanaryGitOID{}, errors.New("soak canary detached Git HEAD is invalid")
	}
	return oid, nil
}

func resolveSoakCanaryGitReference(
	layout soakCanaryGitLayout,
	name string,
	depth int,
) (soakCanaryGitOID, error) {
	if depth >= 8 || !validSoakCanaryGitReferenceName(name) {
		return soakCanaryGitOID{}, errors.New("soak canary Git reference is invalid")
	}
	for _, directory := range []string{layout.gitDir, layout.commonDir} {
		path := filepath.Join(directory, filepath.FromSlash(name))
		body, err := readBoundedRegularFile(path, soakCanaryGitMaxReferenceLen)
		if err == nil {
			value := strings.TrimSpace(string(body))
			if strings.HasPrefix(value, "ref: ") {
				return resolveSoakCanaryGitReference(
					layout,
					strings.TrimSpace(strings.TrimPrefix(value, "ref: ")),
					depth+1,
				)
			}
			oid, parseErr := parseSoakCanaryGitOID(value)
			if parseErr != nil {
				return soakCanaryGitOID{}, errors.New("soak canary loose Git reference is invalid")
			}
			return oid, nil
		}
		if !os.IsNotExist(err) {
			return soakCanaryGitOID{}, err
		}
		if directory == layout.commonDir {
			break
		}
	}
	body, err := readBoundedRegularFile(
		filepath.Join(layout.commonDir, "packed-refs"),
		soakCanaryGitMaxIndexBytes,
	)
	if err != nil {
		return soakCanaryGitOID{}, fmt.Errorf("resolve soak canary Git reference %s: %w", name, err)
	}
	for _, line := range strings.Split(string(body), "\n") {
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "^") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			return soakCanaryGitOID{}, errors.New("soak canary packed Git references are malformed")
		}
		if fields[1] == name {
			oid, parseErr := parseSoakCanaryGitOID(fields[0])
			if parseErr != nil {
				return soakCanaryGitOID{}, errors.New("soak canary packed Git reference is invalid")
			}
			return oid, nil
		}
	}
	return soakCanaryGitOID{}, fmt.Errorf("soak canary Git reference %s is missing", name)
}

func validSoakCanaryGitReferenceName(name string) bool {
	return strings.HasPrefix(name, "refs/") &&
		name == filepath.ToSlash(filepath.Clean(filepath.FromSlash(name))) &&
		!strings.Contains(name, "..") &&
		!strings.ContainsAny(name, "\x00\r\n\\")
}

func loadSoakCanaryGitIndex(layout soakCanaryGitLayout) ([]soakCanaryGitIndexEntry, error) {
	body, err := readBoundedRegularFile(
		filepath.Join(layout.gitDir, "index"),
		soakCanaryGitMaxIndexBytes,
	)
	if err != nil {
		return nil, fmt.Errorf("read soak canary Git index: %w", err)
	}
	if len(body) < 12+sha1.Size || !bytes.Equal(body[:4], []byte("DIRC")) {
		return nil, errors.New("soak canary Git index header is corrupt")
	}
	version := binary.BigEndian.Uint32(body[4:8])
	if version != 2 && version != 3 {
		return nil, fmt.Errorf("soak canary Git index version %d is unsupported", version)
	}
	checksumStart := len(body) - sha1.Size
	checksum := sha1.Sum(body[:checksumStart])
	if !bytes.Equal(checksum[:], body[checksumStart:]) {
		return nil, errors.New("soak canary Git index checksum mismatch")
	}
	count := int(binary.BigEndian.Uint32(body[8:12]))
	if count < 0 || count > soakCanaryGitMaxEntries {
		return nil, errors.New("soak canary Git index entry count exceeds limit")
	}
	entries := make([]soakCanaryGitIndexEntry, 0, count)
	offset := 12
	previousPath := ""
	for index := 0; index < count; index++ {
		entryStart := offset
		if offset+62 > checksumStart {
			return nil, errors.New("soak canary Git index entry is truncated")
		}
		mode := binary.BigEndian.Uint32(body[offset+24 : offset+28])
		var oid soakCanaryGitOID
		copy(oid[:], body[offset+40:offset+60])
		flags := binary.BigEndian.Uint16(body[offset+60 : offset+62])
		offset += 62
		if flags&0x4000 != 0 {
			return nil, errors.New("soak canary Git index extended entries are unsupported")
		}
		if flags&0x3000 != 0 {
			return nil, errors.New("soak canary Git index contains an unmerged stage")
		}
		nameEnd := bytes.IndexByte(body[offset:checksumStart], 0)
		if nameEnd < 0 {
			return nil, errors.New("soak canary Git index path is unterminated")
		}
		pathBytes := body[offset : offset+nameEnd]
		if flags&0x0fff != 0x0fff && int(flags&0x0fff) != len(pathBytes) {
			return nil, errors.New("soak canary Git index path length is inconsistent")
		}
		path := string(pathBytes)
		if !validSoakCanaryGitPath(path) || (previousPath != "" && path <= previousPath) {
			return nil, errors.New("soak canary Git index path is unsafe or unsorted")
		}
		if !validSoakCanaryGitFileMode(mode) {
			return nil, fmt.Errorf("soak canary Git index mode %o is unsupported", mode)
		}
		previousPath = path
		offset += nameEnd + 1
		entryLength := offset - entryStart
		paddedLength := (entryLength + 7) &^ 7
		if entryStart+paddedLength > checksumStart {
			return nil, errors.New("soak canary Git index padding is truncated")
		}
		for _, padding := range body[offset : entryStart+paddedLength] {
			if padding != 0 {
				return nil, errors.New("soak canary Git index padding is corrupt")
			}
		}
		offset = entryStart + paddedLength
		entries = append(entries, soakCanaryGitIndexEntry{path: path, mode: mode, oid: oid})
	}
	seenTree := false
	for offset < checksumStart {
		if offset+8 > checksumStart {
			return nil, errors.New("soak canary Git index extension is truncated")
		}
		signature := string(body[offset : offset+4])
		size := int(binary.BigEndian.Uint32(body[offset+4 : offset+8]))
		offset += 8
		if size < 0 || offset+size > checksumStart {
			return nil, errors.New("soak canary Git index extension size is invalid")
		}
		if signature != "TREE" || seenTree {
			return nil, fmt.Errorf("soak canary Git index extension %q is unsupported", signature)
		}
		seenTree = true
		offset += size
	}
	return entries, nil
}

func validSoakCanaryGitPath(path string) bool {
	if path == "" || strings.HasPrefix(path, "/") || strings.HasSuffix(path, "/") ||
		strings.ContainsAny(path, "\x00\\") || !utf8.ValidString(path) {
		return false
	}
	for _, component := range strings.Split(path, "/") {
		if component == "" || component == "." || component == ".." || component == ".git" {
			return false
		}
	}
	return filepath.ToSlash(filepath.Clean(filepath.FromSlash(path))) == path
}

func validSoakCanaryGitFileMode(mode uint32) bool {
	return mode == 0o100644 || mode == 0o100755 || mode == 0o120000
}

func loadSoakCanaryGitHEADTree(
	store soakCanaryGitObjectStore,
	head soakCanaryGitOID,
) (map[string]soakCanaryGitTreeEntry, error) {
	commit, err := store.object(head, 0)
	if err != nil {
		return nil, fmt.Errorf("read soak canary Git HEAD commit: %w", err)
	}
	if commit.kind != "commit" {
		return nil, errors.New("soak canary Git HEAD does not identify a commit")
	}
	lineEnd := bytes.IndexByte(commit.data, '\n')
	if lineEnd < 0 || !bytes.HasPrefix(commit.data[:lineEnd], []byte("tree ")) {
		return nil, errors.New("soak canary Git HEAD commit has no tree")
	}
	treeOID, err := parseSoakCanaryGitOID(string(commit.data[len("tree "):lineEnd]))
	if err != nil {
		return nil, errors.New("soak canary Git HEAD tree is invalid")
	}
	entries := map[string]soakCanaryGitTreeEntry{}
	if err := loadSoakCanaryGitTree(store, treeOID, "", entries, 0); err != nil {
		return nil, err
	}
	return entries, nil
}

func loadSoakCanaryGitTree(
	store soakCanaryGitObjectStore,
	oid soakCanaryGitOID,
	prefix string,
	entries map[string]soakCanaryGitTreeEntry,
	depth int,
) error {
	if depth >= soakCanaryGitMaxDeltaDepth || len(entries) > soakCanaryGitMaxEntries {
		return errors.New("soak canary Git tree exceeds bounded limits")
	}
	object, err := store.object(oid, 0)
	if err != nil {
		return fmt.Errorf("read soak canary Git tree %s: %w", oid, err)
	}
	if object.kind != "tree" {
		return errors.New("soak canary Git tree object has wrong type")
	}
	offset := 0
	for offset < len(object.data) {
		space := bytes.IndexByte(object.data[offset:], ' ')
		if space <= 0 {
			return errors.New("soak canary Git tree mode is malformed")
		}
		space += offset
		modeValue, err := strconv.ParseUint(string(object.data[offset:space]), 8, 32)
		if err != nil {
			return errors.New("soak canary Git tree mode is invalid")
		}
		nameStart := space + 1
		nameEnd := bytes.IndexByte(object.data[nameStart:], 0)
		if nameEnd <= 0 {
			return errors.New("soak canary Git tree name is malformed")
		}
		nameEnd += nameStart
		if nameEnd+1+sha1.Size > len(object.data) {
			return errors.New("soak canary Git tree object ID is truncated")
		}
		name := string(object.data[nameStart:nameEnd])
		if strings.Contains(name, "/") || !validSoakCanaryGitPath(name) {
			return errors.New("soak canary Git tree name is unsafe")
		}
		var child soakCanaryGitOID
		copy(child[:], object.data[nameEnd+1:nameEnd+1+sha1.Size])
		path := name
		if prefix != "" {
			path = prefix + "/" + name
		}
		mode := uint32(modeValue)
		switch mode {
		case 0o40000:
			if err := loadSoakCanaryGitTree(store, child, path, entries, depth+1); err != nil {
				return err
			}
		case 0o100644, 0o100755, 0o120000:
			if _, exists := entries[path]; exists {
				return errors.New("soak canary Git tree contains a duplicate path")
			}
			entries[path] = soakCanaryGitTreeEntry{mode: mode, oid: child}
		default:
			return fmt.Errorf("soak canary Git tree mode %o is unsupported", mode)
		}
		offset = nameEnd + 1 + sha1.Size
		if len(entries) > soakCanaryGitMaxEntries {
			return errors.New("soak canary Git tree entry count exceeds limit")
		}
	}
	return nil
}

func compareSoakCanaryGitIndexToHEAD(
	index []soakCanaryGitIndexEntry,
	tree map[string]soakCanaryGitTreeEntry,
) error {
	if len(index) != len(tree) {
		return errors.New("soak canary Git index differs from HEAD")
	}
	for _, entry := range index {
		head, exists := tree[entry.path]
		if !exists || entry.mode != head.mode || entry.oid != head.oid {
			return fmt.Errorf("soak canary Git index differs from HEAD at %s", entry.path)
		}
	}
	return nil
}

func verifySoakCanaryGitWorktree(
	root string,
	index []soakCanaryGitIndexEntry,
	checkExecutableMode bool,
) error {
	for _, entry := range index {
		path := filepath.Join(root, filepath.FromSlash(entry.path))
		info, err := os.Lstat(path)
		if err != nil {
			return fmt.Errorf("soak canary Git worktree path %s is missing", entry.path)
		}
		var oid soakCanaryGitOID
		switch entry.mode {
		case 0o100644, 0o100755:
			if !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("soak canary Git worktree mode mismatch at %s", entry.path)
			}
			if checkExecutableMode && runtime.GOOS != "windows" {
				executable := info.Mode().Perm()&0o111 != 0
				if executable != (entry.mode == 0o100755) {
					return fmt.Errorf("soak canary Git executable mode mismatch at %s", entry.path)
				}
			}
			oid, err = hashSoakCanaryGitRegularBlob(path, info.Size())
		case 0o120000:
			if info.Mode()&os.ModeSymlink == 0 {
				return fmt.Errorf("soak canary Git symlink mode mismatch at %s", entry.path)
			}
			target, readErr := os.Readlink(path)
			if readErr != nil {
				err = readErr
			} else {
				oid = hashSoakCanaryGitBlob([]byte(target))
			}
		}
		if err != nil {
			return fmt.Errorf("hash soak canary Git worktree path %s: %w", entry.path, err)
		}
		if oid != entry.oid {
			return fmt.Errorf("soak canary Git worktree content mismatch at %s", entry.path)
		}
	}
	return nil
}

func rejectSoakCanaryGitUntracked(root string, index []soakCanaryGitIndexEntry) error {
	tracked := make(map[string]bool, len(index))
	seen := make(map[string]bool, len(index))
	for _, entry := range index {
		tracked[entry.path] = true
	}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if relative == "." {
			return nil
		}
		relative = filepath.ToSlash(relative)
		if relative == ".git" {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() && info.Mode()&os.ModeSymlink == 0 {
			return fmt.Errorf("soak canary Git worktree contains unsupported path %s", relative)
		}
		if !tracked[relative] {
			return fmt.Errorf("soak canary Git worktree contains untracked path %s", relative)
		}
		seen[relative] = true
		return nil
	})
	if err != nil {
		return err
	}
	if len(seen) != len(tracked) {
		return errors.New("soak canary Git worktree changed while verifying tracked paths")
	}
	return nil
}

func soakCanaryGitCoreFileMode(layout soakCanaryGitLayout) (bool, error) {
	fileMode := runtime.GOOS != "windows"
	for _, path := range []string{
		filepath.Join(layout.commonDir, "config"),
		filepath.Join(layout.gitDir, "config.worktree"),
	} {
		body, err := readBoundedRegularFile(path, 1<<20)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return false, err
		}
		section := ""
		for _, rawLine := range strings.Split(string(body), "\n") {
			line := strings.TrimSpace(rawLine)
			if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
				continue
			}
			if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
				section = strings.ToLower(strings.TrimSpace(line[1 : len(line)-1]))
				continue
			}
			if section != "core" {
				continue
			}
			key, value, found := strings.Cut(line, "=")
			if !found || !strings.EqualFold(strings.TrimSpace(key), "filemode") {
				continue
			}
			parsed, parseErr := strconv.ParseBool(strings.TrimSpace(value))
			if parseErr != nil {
				return false, errors.New("soak canary Git core.filemode is malformed")
			}
			fileMode = parsed
		}
	}
	return fileMode, nil
}

func hashSoakCanaryGitRegularBlob(path string, size int64) (soakCanaryGitOID, error) {
	if size < 0 || size > soakCanarySnapshotMaximumFileBytes {
		return soakCanaryGitOID{}, errors.New("soak canary Git worktree file exceeds limit")
	}
	file, err := os.Open(path)
	if err != nil {
		return soakCanaryGitOID{}, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() != size {
		return soakCanaryGitOID{}, errors.New("soak canary Git worktree file changed while hashing")
	}
	hasher := sha1.New()
	_, _ = fmt.Fprintf(hasher, "blob %d\x00", size)
	written, err := io.CopyN(hasher, file, size)
	if err != nil || written != size {
		return soakCanaryGitOID{}, errors.New("soak canary Git worktree file is truncated")
	}
	var extra [1]byte
	if count, readErr := file.Read(extra[:]); count != 0 || (readErr != nil && readErr != io.EOF) {
		return soakCanaryGitOID{}, errors.New("soak canary Git worktree file changed while hashing")
	}
	var oid soakCanaryGitOID
	copy(oid[:], hasher.Sum(nil))
	return oid, nil
}

func hashSoakCanaryGitBlob(body []byte) soakCanaryGitOID {
	hasher := sha1.New()
	_, _ = fmt.Fprintf(hasher, "blob %d\x00", len(body))
	_, _ = hasher.Write(body)
	var oid soakCanaryGitOID
	copy(oid[:], hasher.Sum(nil))
	return oid
}

func parseSoakCanaryGitOID(value string) (soakCanaryGitOID, error) {
	var oid soakCanaryGitOID
	if len(value) != hex.EncodedLen(len(oid)) {
		return oid, errors.New("Git object ID has wrong length")
	}
	body, err := hex.DecodeString(value)
	if err != nil {
		return oid, err
	}
	copy(oid[:], body)
	return oid, nil
}

func (oid soakCanaryGitOID) String() string {
	return hex.EncodeToString(oid[:])
}

func (store soakCanaryGitObjectStore) object(
	oid soakCanaryGitOID,
	depth int,
) (soakCanaryGitObject, error) {
	if depth >= soakCanaryGitMaxDeltaDepth {
		return soakCanaryGitObject{}, errors.New("soak canary Git object delta depth exceeds limit")
	}
	object, err := store.looseObject(oid)
	if err == nil {
		return object, nil
	}
	if !os.IsNotExist(err) {
		return soakCanaryGitObject{}, err
	}
	object, err = store.packedObject(oid, depth)
	if err != nil {
		return soakCanaryGitObject{}, err
	}
	if hashSoakCanaryGitObject(object.kind, object.data) != oid {
		return soakCanaryGitObject{}, errors.New("soak canary packed Git object hash mismatch")
	}
	return object, nil
}

func (store soakCanaryGitObjectStore) looseObject(
	oid soakCanaryGitOID,
) (soakCanaryGitObject, error) {
	value := oid.String()
	path := filepath.Join(store.commonDir, "objects", value[:2], value[2:])
	file, err := os.Open(path)
	if err != nil {
		return soakCanaryGitObject{}, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return soakCanaryGitObject{}, errors.New("soak canary loose Git object is unsafe")
	}
	reader, err := zlib.NewReader(file)
	if err != nil {
		return soakCanaryGitObject{}, errors.New("soak canary loose Git object is corrupt")
	}
	defer reader.Close()
	raw, err := io.ReadAll(io.LimitReader(reader, soakCanaryGitMaxObjectBytes+256+1))
	if err != nil || len(raw) > soakCanaryGitMaxObjectBytes+256 {
		return soakCanaryGitObject{}, errors.New("soak canary loose Git object exceeds limit")
	}
	separator := bytes.IndexByte(raw, 0)
	if separator <= 0 {
		return soakCanaryGitObject{}, errors.New("soak canary loose Git object header is corrupt")
	}
	fields := strings.Fields(string(raw[:separator]))
	if len(fields) != 2 {
		return soakCanaryGitObject{}, errors.New("soak canary loose Git object header is invalid")
	}
	size, err := strconv.Atoi(fields[1])
	if err != nil || size != len(raw)-separator-1 || size > soakCanaryGitMaxObjectBytes {
		return soakCanaryGitObject{}, errors.New("soak canary loose Git object size is invalid")
	}
	object := soakCanaryGitObject{kind: fields[0], data: raw[separator+1:]}
	if hashSoakCanaryGitObject(object.kind, object.data) != oid {
		return soakCanaryGitObject{}, errors.New("soak canary loose Git object hash mismatch")
	}
	return object, nil
}

func (store soakCanaryGitObjectStore) packedObject(
	oid soakCanaryGitOID,
	depth int,
) (soakCanaryGitObject, error) {
	packDirectory := filepath.Join(store.commonDir, "objects", "pack")
	children, err := os.ReadDir(packDirectory)
	if err != nil {
		return soakCanaryGitObject{}, fmt.Errorf("read soak canary Git pack directory: %w", err)
	}
	var indexes []string
	for _, child := range children {
		if !child.Type().IsRegular() || !strings.HasSuffix(child.Name(), ".idx") {
			continue
		}
		indexes = append(indexes, filepath.Join(packDirectory, child.Name()))
	}
	sort.Strings(indexes)
	if len(indexes) > soakCanaryGitMaxPacks {
		return soakCanaryGitObject{}, errors.New("soak canary Git pack count exceeds limit")
	}
	for _, indexPath := range indexes {
		packPath, offset, found, err := findSoakCanaryGitPackedOffset(indexPath, oid)
		if err != nil {
			return soakCanaryGitObject{}, err
		}
		if found {
			return store.packObjectAt(packPath, offset, depth, map[int64]bool{})
		}
	}
	return soakCanaryGitObject{}, fmt.Errorf("soak canary Git object %s is missing", oid)
}

func findSoakCanaryGitPackedOffset(
	indexPath string,
	oid soakCanaryGitOID,
) (string, int64, bool, error) {
	body, err := readBoundedRegularFile(indexPath, soakCanaryGitMaxPackIndex)
	if err != nil {
		return "", 0, false, err
	}
	if len(body) < 8+256*4+40 ||
		!bytes.Equal(body[:4], []byte{0xff, 0x74, 0x4f, 0x63}) ||
		binary.BigEndian.Uint32(body[4:8]) != 2 {
		return "", 0, false, errors.New("soak canary Git pack index version is unsupported")
	}
	indexChecksum := sha1.Sum(body[:len(body)-sha1.Size])
	if !bytes.Equal(indexChecksum[:], body[len(body)-sha1.Size:]) {
		return "", 0, false, errors.New("soak canary Git pack index checksum mismatch")
	}
	fanout := body[8 : 8+256*4]
	for index := 1; index < 256; index++ {
		if binary.BigEndian.Uint32(fanout[index*4:(index+1)*4]) <
			binary.BigEndian.Uint32(fanout[(index-1)*4:index*4]) {
			return "", 0, false, errors.New("soak canary Git pack index fanout is corrupt")
		}
	}
	count := int(binary.BigEndian.Uint32(fanout[255*4 : 256*4]))
	if count < 0 || count > soakCanaryGitMaxEntries*10 {
		return "", 0, false, errors.New("soak canary Git pack index entry count exceeds limit")
	}
	namesStart := 8 + 256*4
	crcStart := namesStart + count*sha1.Size
	offsetsStart := crcStart + count*4
	largeStart := offsetsStart + count*4
	if largeStart+40 > len(body) {
		return "", 0, false, errors.New("soak canary Git pack index is truncated")
	}
	for index := 1; index < count; index++ {
		previous := body[namesStart+(index-1)*sha1.Size : namesStart+index*sha1.Size]
		current := body[namesStart+index*sha1.Size : namesStart+(index+1)*sha1.Size]
		if bytes.Compare(previous, current) >= 0 {
			return "", 0, false, errors.New("soak canary Git pack index names are unsorted")
		}
	}
	index := sort.Search(count, func(index int) bool {
		return bytes.Compare(body[namesStart+index*sha1.Size:namesStart+(index+1)*sha1.Size], oid[:]) >= 0
	})
	if index >= count ||
		!bytes.Equal(body[namesStart+index*sha1.Size:namesStart+(index+1)*sha1.Size], oid[:]) {
		return "", 0, false, nil
	}
	value := binary.BigEndian.Uint32(body[offsetsStart+index*4 : offsetsStart+(index+1)*4])
	var offset int64
	if value&0x80000000 == 0 {
		offset = int64(value)
	} else {
		largeIndex := int(value & 0x7fffffff)
		position := largeStart + largeIndex*8
		if position+8 > len(body)-40 {
			return "", 0, false, errors.New("soak canary Git pack large offset is invalid")
		}
		large := binary.BigEndian.Uint64(body[position : position+8])
		if large > uint64(^uint64(0)>>1) {
			return "", 0, false, errors.New("soak canary Git pack offset exceeds limit")
		}
		offset = int64(large)
	}
	packPath := strings.TrimSuffix(indexPath, ".idx") + ".pack"
	pack, err := os.Open(packPath)
	if err != nil {
		return "", 0, false, err
	}
	defer pack.Close()
	info, err := pack.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() < 12+sha1.Size {
		return "", 0, false, errors.New("soak canary Git pack is unsafe")
	}
	var header [12]byte
	if _, err := pack.ReadAt(header[:], 0); err != nil ||
		!bytes.Equal(header[:4], []byte("PACK")) ||
		(binary.BigEndian.Uint32(header[4:8]) != 2 &&
			binary.BigEndian.Uint32(header[4:8]) != 3) {
		return "", 0, false, errors.New("soak canary Git pack header is invalid")
	}
	var trailer [sha1.Size]byte
	if _, err := pack.ReadAt(trailer[:], info.Size()-sha1.Size); err != nil {
		return "", 0, false, err
	}
	if !bytes.Equal(trailer[:], body[len(body)-2*sha1.Size:len(body)-sha1.Size]) {
		return "", 0, false, errors.New("soak canary Git pack checksum binding mismatch")
	}
	if offset < 12 || offset >= info.Size()-sha1.Size {
		return "", 0, false, errors.New("soak canary Git pack offset is invalid")
	}
	return packPath, offset, true, nil
}

func (store soakCanaryGitObjectStore) packObjectAt(
	packPath string,
	objectOffset int64,
	depth int,
	seen map[int64]bool,
) (soakCanaryGitObject, error) {
	if depth >= soakCanaryGitMaxDeltaDepth || seen[objectOffset] {
		return soakCanaryGitObject{}, errors.New("soak canary Git pack delta is cyclic or too deep")
	}
	seen[objectOffset] = true
	defer delete(seen, objectOffset)
	file, err := os.Open(packPath)
	if err != nil {
		return soakCanaryGitObject{}, err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || objectOffset < 12 || objectOffset >= info.Size()-sha1.Size {
		return soakCanaryGitObject{}, errors.New("soak canary Git pack object offset is invalid")
	}
	position := objectOffset
	first, err := readSoakCanaryGitByteAt(file, &position)
	if err != nil {
		return soakCanaryGitObject{}, err
	}
	objectType := (first >> 4) & 7
	size := int64(first & 0x0f)
	shift := uint(4)
	for first&0x80 != 0 {
		first, err = readSoakCanaryGitByteAt(file, &position)
		if err != nil {
			return soakCanaryGitObject{}, err
		}
		if shift >= 63 || int64(first&0x7f) > (soakCanaryGitMaxObjectBytes>>shift) {
			return soakCanaryGitObject{}, errors.New("soak canary Git pack object size exceeds limit")
		}
		size |= int64(first&0x7f) << shift
		shift += 7
	}
	if size < 0 || size > soakCanaryGitMaxObjectBytes {
		return soakCanaryGitObject{}, errors.New("soak canary Git pack object size exceeds limit")
	}
	var baseOffset int64
	var baseOID soakCanaryGitOID
	switch objectType {
	case 6:
		value, err := readSoakCanaryGitByteAt(file, &position)
		if err != nil {
			return soakCanaryGitObject{}, err
		}
		distance := int64(value & 0x7f)
		for value&0x80 != 0 {
			value, err = readSoakCanaryGitByteAt(file, &position)
			if err != nil {
				return soakCanaryGitObject{}, err
			}
			if distance > (1<<56)-1 {
				return soakCanaryGitObject{}, errors.New("soak canary Git pack delta offset exceeds limit")
			}
			distance = ((distance + 1) << 7) | int64(value&0x7f)
		}
		baseOffset = objectOffset - distance
		if baseOffset < 12 {
			return soakCanaryGitObject{}, errors.New("soak canary Git pack delta base offset is invalid")
		}
	case 7:
		if _, err := file.ReadAt(baseOID[:], position); err != nil {
			return soakCanaryGitObject{}, err
		}
		position += sha1.Size
	case 1, 2, 3, 4:
	default:
		return soakCanaryGitObject{}, errors.New("soak canary Git pack object type is unsupported")
	}
	reader, err := zlib.NewReader(io.NewSectionReader(
		file,
		position,
		info.Size()-sha1.Size-position,
	))
	if err != nil {
		return soakCanaryGitObject{}, errors.New("soak canary Git pack object compression is corrupt")
	}
	data, readErr := io.ReadAll(io.LimitReader(reader, soakCanaryGitMaxObjectBytes+1))
	closeErr := reader.Close()
	if readErr != nil || closeErr != nil || int64(len(data)) != size {
		return soakCanaryGitObject{}, errors.New("soak canary Git pack object data is corrupt")
	}
	switch objectType {
	case 1, 2, 3, 4:
		kinds := map[byte]string{1: "commit", 2: "tree", 3: "blob", 4: "tag"}
		return soakCanaryGitObject{kind: kinds[objectType], data: data}, nil
	case 6, 7:
		var base soakCanaryGitObject
		if objectType == 6 {
			base, err = store.packObjectAt(packPath, baseOffset, depth+1, seen)
		} else {
			base, err = store.object(baseOID, depth+1)
		}
		if err != nil {
			return soakCanaryGitObject{}, err
		}
		result, err := applySoakCanaryGitDelta(base.data, data)
		if err != nil {
			return soakCanaryGitObject{}, err
		}
		return soakCanaryGitObject{kind: base.kind, data: result}, nil
	}
	return soakCanaryGitObject{}, errors.New("unreachable soak canary Git pack object type")
}

func readSoakCanaryGitByteAt(file *os.File, offset *int64) (byte, error) {
	var body [1]byte
	if _, err := file.ReadAt(body[:], *offset); err != nil {
		return 0, err
	}
	*offset = *offset + 1
	return body[0], nil
}

func applySoakCanaryGitDelta(base, delta []byte) ([]byte, error) {
	baseSize, offset, err := readSoakCanaryGitDeltaSize(delta, 0)
	if err != nil || baseSize != int64(len(base)) {
		return nil, errors.New("soak canary Git delta base size is invalid")
	}
	resultSize, offset, err := readSoakCanaryGitDeltaSize(delta, offset)
	if err != nil || resultSize < 0 || resultSize > soakCanaryGitMaxObjectBytes {
		return nil, errors.New("soak canary Git delta result size is invalid")
	}
	result := make([]byte, 0, resultSize)
	for offset < len(delta) {
		command := delta[offset]
		offset++
		if command&0x80 == 0 {
			count := int(command)
			if count == 0 || offset+count > len(delta) ||
				int64(len(result)+count) > resultSize {
				return nil, errors.New("soak canary Git delta literal is invalid")
			}
			result = append(result, delta[offset:offset+count]...)
			offset += count
			continue
		}
		copyOffset := 0
		copySize := 0
		for bit := byte(0); bit < 4; bit++ {
			if command&(1<<bit) != 0 {
				if offset >= len(delta) {
					return nil, errors.New("soak canary Git delta copy offset is truncated")
				}
				copyOffset |= int(delta[offset]) << (8 * bit)
				offset++
			}
		}
		for bit := byte(0); bit < 3; bit++ {
			if command&(1<<(4+bit)) != 0 {
				if offset >= len(delta) {
					return nil, errors.New("soak canary Git delta copy size is truncated")
				}
				copySize |= int(delta[offset]) << (8 * bit)
				offset++
			}
		}
		if copySize == 0 {
			copySize = 0x10000
		}
		if copyOffset < 0 || copySize < 0 || copyOffset+copySize > len(base) ||
			int64(len(result)+copySize) > resultSize {
			return nil, errors.New("soak canary Git delta copy is out of bounds")
		}
		result = append(result, base[copyOffset:copyOffset+copySize]...)
	}
	if int64(len(result)) != resultSize {
		return nil, errors.New("soak canary Git delta result length mismatch")
	}
	return result, nil
}

func readSoakCanaryGitDeltaSize(body []byte, offset int) (int64, int, error) {
	var value int64
	var shift uint
	for {
		if offset >= len(body) || shift >= 63 {
			return 0, offset, errors.New("soak canary Git delta size is truncated")
		}
		current := body[offset]
		offset++
		value |= int64(current&0x7f) << shift
		if current&0x80 == 0 {
			return value, offset, nil
		}
		shift += 7
	}
}

func hashSoakCanaryGitObject(kind string, data []byte) soakCanaryGitOID {
	hasher := sha1.New()
	_, _ = fmt.Fprintf(hasher, "%s %d\x00", kind, len(data))
	_, _ = hasher.Write(data)
	var oid soakCanaryGitOID
	copy(oid[:], hasher.Sum(nil))
	return oid
}
