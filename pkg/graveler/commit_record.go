package graveler

func CommitDataToCommitRecord(c *CommitData) *CommitRecord {
	var parents []CommitID
	for _, parent := range c.Parents {
		parents = append(parents, CommitID(parent))
	}

	return &CommitRecord{
		CommitID: CommitID(c.Id),
		Commit: &Commit{
			Committer:    c.Committer,
			Message:      c.Message,
			CreationDate: c.CreationDate.AsTime(),
			MetaRangeID:  MetaRangeID(c.MetaRangeId),
			Metadata:     c.Metadata,
			Parents:      parents,
			Version:      CommitVersion(c.Version),
			Generation:   CommitGeneration(c.Generation),
		},
	}
}
