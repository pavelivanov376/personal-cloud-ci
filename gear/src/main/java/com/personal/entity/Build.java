package com.personal.entity;

import jakarta.persistence.*;
import java.time.Instant;

@Entity
@Table(name = "build_entity")
public class Build {
    @Id
    private String id;
    private int buildNumber;
    private Instant timestamp;
    private String status;
    private String repositoryUrl;

    public String getId() { return id; }
    public int getBuildNumber() { return buildNumber; }
    public Instant getTimestamp() { return timestamp; }
    public String getStatus() { return status; }
    public String getRepositoryUrl() { return repositoryUrl; }
}
