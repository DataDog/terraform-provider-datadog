package cassette

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/terraform-providers/terraform-provider-datadog/generator/internal/model"
)

// ----------------------------------------------------------------------------
// Bundle writing
//
// A cassette bundle is three files that only mean anything together: the
// generated test, the fixture it replays, and the freeze companion that fixes
// the clock. Writing some of them is worse than writing none — a test without
// its cassette fails as a confusing replay error rather than an obvious
// absence — so the three are rendered and classified before any path is
// touched, and a failure part-way restores what was there.
// ----------------------------------------------------------------------------

// WritePolicy decides what an existing bundle on disk authorizes.
type WritePolicy string

// WritePolicyMissing is the default: write only when no member exists.
// Replacing tfgen's own earlier output, or a hand-recorded fixture, are
// separate authorizations a maintainer has to ask for explicitly.
const WritePolicyMissing WritePolicy = "missing"

// BundleCollisionError reports a bundle the policy does not authorize writing.
// It names the ownership found and the members responsible, because the way
// out differs: regenerating over tfgen's own output is routine, while
// replacing a hand-recorded cassette is a decision about someone else's work.
type BundleCollisionError struct {
	Ownership model.CassetteOwnership
	// Existing are the members already on disk, in bundle order.
	Existing []string
	Policy   WritePolicy
}

func (e *BundleCollisionError) Error() string {
	switch e.Ownership {
	case model.CassetteOwnershipIncomplete:
		return fmt.Sprintf(
			"refusing to write a cassette bundle over a partial one: %s already exist(s) while its "+
				"siblings do not, so neither tfgen's ownership nor a maintainer's intent can be read from it. "+
				"Remove the stray file, or rename the artifact",
			strings.Join(e.Existing, ", "))
	case model.CassetteOwnershipHandwritten:
		return fmt.Sprintf(
			"refusing to replace a hand-recorded cassette bundle: %s carr(ies) no %q marker, so it was "+
				"recorded against a real org rather than generated. A recording is evidence the API behaved "+
				"a certain way, which a generated fixture is not, so replacing it takes an explicit policy",
			strings.Join(e.Existing, ", "), model.GeneratedMarker)
	default:
		return fmt.Sprintf(
			"refusing to replace an existing generated cassette bundle under the %q policy: %s already "+
				"exist(s). Regeneration is authorized separately so an unreviewed fixture change cannot "+
				"ride along with an unrelated run",
			e.Policy, strings.Join(e.Existing, ", "))
	}
}

// WriteBundle writes a rendered bundle under the given policy, reporting what
// it did.
//
// Nothing is written unless all three members were rendered: a bundle that
// cannot be completed is reported rather than partially committed. The on-disk
// ownership is classified first, and under the default policy only a wholly
// absent bundle is written.
func WriteBundle(bundle *model.CassetteBundle, policy WritePolicy) (model.CassetteWriteAction, error) {
	if !bundle.Complete() {
		return model.CassetteWriteNone, fmt.Errorf(
			"refusing to write an incomplete cassette bundle: all three of the test, cassette and freeze " +
				"companion must be rendered before any is committed")
	}

	ownership, existing, err := classifyBundle(bundle)
	if err != nil {
		return model.CassetteWriteNone, err
	}
	bundle.Ownership = ownership

	if ownership != model.CassetteOwnershipMissing {
		return model.CassetteWriteNone, &BundleCollisionError{
			Ownership: ownership, Existing: existing, Policy: policy,
		}
	}

	if err := commit(bundle); err != nil {
		return model.CassetteWriteNone, err
	}
	return model.CassetteWriteCreated, nil
}

// ----------------------------------------------------------------------------
// Ownership preflight
// ----------------------------------------------------------------------------

// classifyBundle reports what is on disk and which members account for it.
//
// Only the test and the cassette carry a marker; the freeze companion is a
// bare timestamp, so its ownership derives from the cassette beside it. A
// bundle whose members disagree is Incomplete rather than resolved one way:
// guessing would either clobber a recording or refuse tfgen's own output.
func classifyBundle(bundle *model.CassetteBundle) (model.CassetteOwnership, []string, error) {
	var existing []string
	marked := 0
	markable := 0

	for _, path := range bundle.Paths() {
		content, err := readIfExists(path)
		if err != nil {
			return "", nil, err
		}
		if content == nil {
			continue
		}
		existing = append(existing, path)
		if path == bundle.FreezePath {
			// No marker to read; it follows the cassette.
			continue
		}
		markable++
		if bytes.Contains(content, []byte(model.GeneratedMarker)) {
			marked++
		}
	}

	switch {
	case len(existing) == 0:
		return model.CassetteOwnershipMissing, nil, nil
	case len(existing) != len(bundle.Paths()):
		return model.CassetteOwnershipIncomplete, existing, nil
	case markable > 0 && marked == markable:
		return model.CassetteOwnershipGenerated, existing, nil
	case marked == 0:
		return model.CassetteOwnershipHandwritten, existing, nil
	default:
		// Some members claim tfgen's ownership and some do not, which is
		// neither a recording nor tfgen's own output.
		return model.CassetteOwnershipIncomplete, existing, nil
	}
}

// readIfExists returns a file's content, or nil when it is absent.
func readIfExists(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return content, nil
}

// ----------------------------------------------------------------------------
// Atomic commit
// ----------------------------------------------------------------------------

// member pairs one bundle path with the bytes destined for it.
type member struct {
	path    string
	content []byte
}

// commit writes every member, restoring the previous state if any write fails.
//
// Each member lands via a temporary file renamed into place, so a reader never
// observes a half-written fixture. The rollback undoes whichever members
// already landed: this matters even under the missing-only policy, where
// "restore" means remove, because a partially written bundle would be
// classified Incomplete on the next run and block regeneration until someone
// cleaned it up by hand.
func commit(bundle *model.CassetteBundle) error {
	members := []member{
		{bundle.TestPath, bundle.TestContent},
		{bundle.CassettePath, bundle.CassetteContent},
		{bundle.FreezePath, bundle.FreezeContent},
	}

	var written []string
	rollback := func() {
		// Reverse order, so a directory created for the first member is
		// removed last if it ends up empty.
		for _, path := range slices.Backward(written) {
			_ = os.Remove(path)
		}
	}

	for _, m := range members {
		if err := os.MkdirAll(filepath.Dir(m.path), 0o755); err != nil {
			rollback()
			return fmt.Errorf("creating directory for %s: %w", m.path, err)
		}
		if err := writeAtomic(m.path, m.content); err != nil {
			rollback()
			return err
		}
		written = append(written, m.path)
	}
	return nil
}

// writeAtomic writes content to path through a temporary file in the same
// directory, so the rename is atomic on the same filesystem and an interrupted
// write leaves the target untouched.
func writeAtomic(path string, content []byte) error {
	dir := filepath.Dir(path)
	temp, err := os.CreateTemp(dir, filepath.Base(path)+".tfgen-*")
	if err != nil {
		return fmt.Errorf("creating temporary file for %s: %w", path, err)
	}
	tempPath := temp.Name()

	cleanup := func() {
		_ = temp.Close()
		_ = os.Remove(tempPath)
	}
	if _, err := temp.Write(content); err != nil {
		cleanup()
		return fmt.Errorf("writing %s: %w", path, err)
	}
	if err := temp.Close(); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("closing temporary file for %s: %w", path, err)
	}
	// Committed files are read by tooling and by humans, so they get the
	// repository's ordinary file mode rather than a temp file's 0600.
	if err := os.Chmod(tempPath, 0o644); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("setting mode on %s: %w", path, err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("renaming into %s: %w", path, err)
	}
	return nil
}
