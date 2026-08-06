package com.personal.entities;

import jakarta.persistence.*;
import java.time.Instant;

@Entity
public class BuildEntity {
    @Id @GeneratedValue(strategy = GenerationType.UUID)
    private String id;
    private int buildNumber;
    private Instant timestamp;
    private String status;
    private String repositoryUrl;

    @ManyToOne
    @JoinColumn(name = "job_id")
    private JobEntity job;

    public String getId() { return id; }
    public int getBuildNumber() { return buildNumber; }
    public void setBuildNumber(int buildNumber) { this.buildNumber = buildNumber; }
    public Instant getTimestamp() { return timestamp; }
    public void setTimestamp(Instant timestamp) { this.timestamp = timestamp; }
    public String getStatus() { return status; }
    public void setStatus(String status) { this.status = status; }
    public String getRepositoryUrl() { return repositoryUrl; }
    public void setRepositoryUrl(String repositoryUrl) { this.repositoryUrl = repositoryUrl; }
    public JobEntity getJob() { return job; }
    public void setJob(JobEntity job) { this.job = job; }
}
