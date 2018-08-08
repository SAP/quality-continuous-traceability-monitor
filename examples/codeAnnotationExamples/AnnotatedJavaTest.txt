package com.myCompany.myapp;

// Tracing entire test class to requirements GitHub#1 and Jira#1, Jira#2
// Trace(GitHub:myOrg/mySourcecodeRepo#1, Jira:MYJIRAPROJECT-1, Jira:MYJIRAPROJECT-2)
public AnnotatedJavaTest {

    @Test
    public void aTestMethodThatIsNotTraced() {
        // assertThat(..., is(...));
    }
    
    // Tracing test method to requirement Jira#3	
    // Trace(Jira:MYJIRAPROJECT-3)
    @Test
    public void aTestMethodThatIsTraced() {
        // assertThat(..., is(...));
    }
}

