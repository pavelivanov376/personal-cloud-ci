package com.personal.entities;

import jakarta.persistence.*;

@Entity
public class BuildEntity {
    @Id @GeneratedValue(strategy = GenerationType.UUID)
    private String id;
    private String status;
    private String repositoryUrl;

    @ManyToOne
    @JoinColumn(name = "job_id")
    private JobEntity job;

    public String getId() { return id; }
    public String getStatus() { return status; }
    public void setStatus(String status) { this.status = status; }
    public String getRepositoryUrl() { return repositoryUrl; }
    public void setRepositoryUrl(String repositoryUrl) { this.repositoryUrl = repositoryUrl; }
    public JobEntity getJob() { return job; }
    public void setJob(JobEntity job) { this.job = job; }
}
