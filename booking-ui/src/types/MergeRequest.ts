export default class MergeRequest {
  id: string = "";
  email: string = "";
  userId: string = "";

  constructor(id: string, email: string, userId: string) {
    this.id = id;
    this.email = email;
    this.userId = userId;
  }
}
